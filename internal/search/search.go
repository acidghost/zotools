// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package search

import (
	"flag"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"github.com/acidghost/zotools/internal/common"
	"github.com/acidghost/zotools/internal/zotero"
	"github.com/fatih/color"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const searchUsageTop = " " + common.OptionsUsage + " search-regexp"

const searchUsageBottom = `  search-regexp
        regexp to search for in the library (case-insensitive, in title)
`

var numCPU = runtime.NumCPU()

var (
	titleColor  = color.New(color.FgGreen, color.Bold)
	selColor    = color.New(color.FgMagenta)
	attachColor = color.New(color.FgBlue)
)

type Command struct {
	fs           *flag.FlagSet
	flagAbstract *bool
	flagAuthors  *bool
	flagSens     *bool
	flagPar      *uint
}

func New(cmd, banner string) *Command {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagAbstract := fs.Bool("abs", false, "search also in the abstract")
	flagAuthors := fs.Bool("auth", false, "search also among the authors")
	flagSens := fs.Bool("s", false, "regular expression is case sensitive")
	flagPar := fs.Uint("j", uint(numCPU),
		fmt.Sprintf("number of search jobs (between 1 and %d)", numCPU))
	fs.Usage = common.MakeUsage(fs, cmd, banner, searchUsageTop, searchUsageBottom)
	return &Command{fs, flagAbstract, flagAuthors, flagSens, flagPar}
}

func (c *Command) Run(args []string, config common.Config) {
	//nolint:errcheck
	c.fs.Parse(args)
	search := c.fs.Arg(0)
	if search == "" {
		c.fs.Usage()
		common.Quit(1)
	}

	par := int(*c.flagPar)
	if par < 1 || par > numCPU {
		common.Die("Number of jobs must be between 1 and %d\n", numCPU)
	}

	var reFlags string
	if !*c.flagSens {
		reFlags += "i"
	}
	if reFlags != "" {
		reFlags = "(?" + reFlags + ")"
	}
	re, err := regexp.Compile(reFlags + search)
	if err != nil {
		common.Die("Wrong search: %v\n", err)
	}

	storage := common.NewStorage(config.Storage)
	if err := storage.Load(); err != nil {
		common.Die("Failed to load the local storage:\n - %v\n", err)
	}

	fmt.Printf("Loaded storage, version %d, %d items\n",
		storage.Data.Lib.Version, len(storage.Data.Lib.Items))

	wgMatchers := sync.WaitGroup{}
	itemsCh := make(chan common.Item)
	matchedCh := make(chan common.Item)
	resCh := make(chan common.SearchResults)

	// Start matcher jobs
	for i := 0; i < par; i++ {
		go func() {
			s := newMatcher(re)
			for item := range itemsCh {
				if c.matchItem(&s, &item) {
					matchedCh <- item
				}
			}
			// No more items to match
			wgMatchers.Done()
		}()
	}
	wgMatchers.Add(par)

	// Wait for the matchers to complete and let the printer job know
	go func() {
		wgMatchers.Wait()
		close(matchedCh)
	}()

	// Printer job, receives from the matchers
	go func() {
		res := common.SearchResults{Term: search, Items: make([]common.SearchResultsItem, 0, 10)}
		var i uint
		for item := range matchedCh {
			titleColor.Print(item.Title)
			if len(item.Creators) > 0 {
				fmt.Printf(" (%s)\n", authorsToString(item.Creators))
			} else {
				fmt.Println()
			}
			for _, attach := range item.Attachments {
				path := common.MakePath(config.Zotero, attach.Key, attach.Filename)
				ns := fmt.Sprintf("%3d)", i)
				fmt.Printf("%s %s\n", selColor.Sprint(ns), attachColor.Sprint(path))
				res.Items = append(res.Items, common.SearchResultsItem{
					Key:         attach.Key,
					Filename:    attach.Filename,
					ContentType: attach.ContentType,
				})
				i++
			}
		}
		resCh <- res
	}()

	// Send all items to matchers
	for _, item := range storage.Data.Lib.Items {
		itemsCh <- item
	}

	// Close to let the matchers know that we're done with the items
	close(itemsCh)
	// Wait for printer to be done
	res := <-resCh
	storage.Data.Search = &res
	if err := storage.Persist(); err != nil {
		common.Die("Failed to persist search:\n - %v\n", err)
	}

	if len(storage.Data.Search.Items) == 0 {
		common.Quit(1)
	}
}

func (c *Command) matchItem(m *matcher, item *common.Item) bool {
	match := m.match(item.Title)
	if *c.flagAbstract {
		if *c.flagAuthors {
			match = match || m.match(item.Abstract) || m.matchAuthors(item.Creators)
		} else {
			match = match || m.match(item.Abstract)
		}
	} else if *c.flagAuthors {
		match = match || m.matchAuthors(item.Creators)
	}
	return match
}

type matcher struct {
	re *regexp.Regexp
	tr *transform.Transformer
}

func newMatcher(re *regexp.Regexp) matcher {
	tr := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)),
		runes.Map(func(r rune) rune {
			switch r {
			case 'Ø':
				return 'O'
			case 'ø':
				return 'o'
			case 'Ł':
				return 'L'
			case 'ł':
				return 'l'
			default:
				return r
			}
		}))
	return matcher{re, &tr}
}

func (m *matcher) match(content string) bool {
	simp, _, _ := transform.String(*m.tr, content)
	return m.re.MatchString(simp)
}

func (m *matcher) matchAuthors(authors []zotero.Creator) bool {
	for _, author := range authors {
		if m.match(author.FirstName) || m.match(author.LastName) {
			return true
		}
	}
	return false
}

func authorsToString(authors []zotero.Creator) string {
	names := make([]string, 0, len(authors))
	for _, author := range authors {
		var initials string
		if len(author.FirstName) > 0 {
			initials = authorInitials(author.FirstName) + " "
		}
		names = append(names, initials+author.LastName)
	}
	return strings.Join(names, ", ")
}

func authorInitials(name string) string {
	initials := make([]string, 0)
	for _, n := range strings.Split(name, " ") {
		initials = append(initials, n[:1]+".")
	}
	return strings.Join(initials, " ")
}
