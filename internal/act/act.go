// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package act

import (
	"flag"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"strings"

	"github.com/acidghost/zotools/internal/common"
	"github.com/mattn/go-shellwords"
)

const actUsageTop = " " + common.OptionsUsage + " [cmd [arg...]]"

const actUsageBottom = `  cmd
        command and arguments to execute
`

type Command struct {
	fs         *flag.FlagSet
	flagIdx    *uint
	flagForget *bool
}

func New(cmd, banner string) *Command {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagIdx := fs.Uint("i", 0, "index into the search results")
	flagForget := fs.Bool("forget", false, "forget previous search")
	fs.Usage = common.MakeUsage(fs, cmd, banner, actUsageTop, actUsageBottom)
	return &Command{fs, flagIdx, flagForget}
}

func (c *Command) Run(args []string, config common.Config) {
	//nolint:errcheck
	c.fs.Parse(args)

	storage := common.NewStorage(config.Storage)
	if err := storage.Load(); err != nil {
		common.Die("Failed to load storage:\n - %v\n", err)
	}

	search := storage.Data.Search

	if *c.flagForget {
		if search != nil {
			storage.Data.Search = nil
			if err := storage.Persist(); err != nil {
				common.Die("Failed to forget search:\n - %v\n", err)
			}
		}
		return
	}

	if search == nil {
		common.Die("No stored search\n")
	} else if int(*c.flagIdx) >= len(search.Items) {
		common.Die("Index %d is invalid: search contains %d items\n",
			*c.flagIdx, len(search.Items))
	}

	item := search.Items[*c.flagIdx]
	path := common.MakePath(config.Zotero, item.Key, item.Filename)
	fmt.Println(path)

	var cmdName string
	var cmdArgs []string
	if c.fs.NArg() == 0 {
		extensions, err := mime.ExtensionsByType(item.ContentType)
		if err != nil {
			common.Die("Could not parse MIME type: %v\n", err)
		} else if extensions == nil {
			common.Die("Unknown extension for MIME type '%s'\n", item.ContentType)
		}
		for _, extension := range extensions {
			varName := "ZOTOOLS_" + strings.ToUpper(extension[1:])
			env := os.Getenv(varName)
			if env != "" {
				envArgs, err := shellwords.Parse(env)
				if err != nil {
					common.Die("Failed to parse %s: %v\n", varName, err)
				}
				cmdName = envArgs[0]
				//nolint:gocritic
				cmdArgs = append(envArgs[1:], path)
				break
			}
		}
		if cmdName == "" {
			common.Die("Command not found for MIME type '%s'\n", item.ContentType)
		}
	} else {
		args = c.fs.Args()
		cmdName = args[0]
		//nolint:gocritic
		cmdArgs = append(args[1:], path)
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		common.Die("Failed to run action: %v\n", err)
	}
}
