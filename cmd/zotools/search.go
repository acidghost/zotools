package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	. "github.com/acidghost/zotools/internal/cache"
	"github.com/acidghost/zotools/internal/zotero"
	"github.com/fatih/color"
)

const SearchUsageFmt = `Usage: %s [OPTIONS] search [OPTIONS] search-regexp

Common options:
`

const SearchUsageFmtOpts = `
Command specific arguments and options:
  search-regexp
        regexp to search for in the library (case-insensitive, in title)
`

type SearchCommand struct {
	fs           *flag.FlagSet
	flagAbstract *bool
	flagAuthors  *bool
}

func NewSearchCommand() *SearchCommand {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	flagAbstract := fs.Bool("abs", false, "search also in the abstract")
	flagAuthors := fs.Bool("auth", false, "search also among the authors")
	fs.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), green(Banner)+"\n\n"+SearchUsageFmt, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), SearchUsageFmtOpts)
		fs.PrintDefaults()
	}
	return &SearchCommand{fs, flagAbstract, flagAuthors}
}

func (c *SearchCommand) Name() string {
	return c.fs.Name()
}

func (c *SearchCommand) Run(args []string, config Config) {
	c.fs.Parse(args)
	search := c.fs.Arg(0)
	if search == "" {
		c.fs.Usage()
		os.Exit(1)
	}

	db, err := LoadDB(config.Cache)
	if err != nil {
		Dief("Failed to load the local database:\n - %v\n", err)
	}

	fmt.Printf("Loaded DB, version %d, %d items\n", db.Lib.Version, len(db.Lib.Items))
	fmt.Printf("Running search for term '%s'\n", blue(search))

	re := regexp.MustCompile("(?i)" + search)
	for _, item := range db.Lib.Items {
		match := re.MatchString(item.Title)
		if *c.flagAbstract {
			if *c.flagAuthors {
				match = match || re.MatchString(item.Abstract) || matchAuthors(re, item.Creators)
			} else {
				match = match || re.MatchString(item.Abstract)
			}
		} else if *c.flagAuthors {
			match = match || matchAuthors(re, item.Creators)
		}
		if match {
			fmt.Printf("%s (%s)\n", bold(green(item.Title)), authorsToString(item.Creators))
			for _, attach := range item.Attachments {
				color.Blue("%s/storage/%s/%s\n", config.Zotero, attach.Key, attach.Filename)
			}
		}
	}
}

func matchAuthors(re *regexp.Regexp, authors []zotero.Creator) bool {
	for _, author := range authors {
		if re.MatchString(author.FirstName) || re.MatchString(author.LastName) {
			return true
		}
	}
	return false
}

func authorsToString(authors []zotero.Creator) string {
	names := make([]string, 0, len(authors))
	for _, author := range authors {
		names = append(names, fmt.Sprintf("%s %s", author.FirstName, author.LastName))
	}
	return strings.Join(names, ", ")
}
