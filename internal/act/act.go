// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU GPL License version 3.

package act

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	. "github.com/acidghost/zotools/internal/common"
)

const actUsageTop = " " + OptionsUsage + " [cmd [arg...]]"

const actUsageBottom = `  cmd
        command and arguments to execute
`

type ActCommand struct {
	fs         *flag.FlagSet
	flagIdx    *uint
	flagForget *bool
}

func New(cmd, banner string) *ActCommand {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagIdx := fs.Uint("i", 0, "index into the search results")
	flagForget := fs.Bool("forget", false, "forget previous search")
	fs.Usage = MakeUsage(fs, cmd, banner, actUsageTop, actUsageBottom)
	return &ActCommand{fs, flagIdx, flagForget}
}

func (c *ActCommand) Run(args []string, config Config) {
	//nolint:errcheck
	c.fs.Parse(args)

	storage := NewStorage(config.Storage)
	if err := storage.Load(); err != nil {
		Dief("Failed to load storage:\n - %v\n", err)
	}

	search := storage.Data.Search

	if *c.flagForget {
		if search != nil {
			storage.Data.Search = nil
			if err := storage.Persist(); err != nil {
				Dief("Failed to forget search:\n - %v\n", err)
			}
		}
		return
	}

	if search == nil {
		Dief("No stored search\n")
	} else if int(*c.flagIdx) >= len(search.Items) {
		Dief("Index %d is invalid: search contains %d items\n", *c.flagIdx, len(search.Items))
	} else if c.fs.NArg() == 0 {
		Dief("Command is missing\n")
	}

	item := search.Items[*c.flagIdx]
	path := MakePath(config.Zotero, item.Key, item.Filename)
	fmt.Println(path)

	args = c.fs.Args()
	cmdName := args[0]
	cmdArgs := append(args[1:], path)
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		Dief("Failed to run action: %v\n", err)
	}
}
