// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/acidghost/zotools/internal/act"
	. "github.com/acidghost/zotools/internal/common"
	"github.com/acidghost/zotools/internal/search"
	"github.com/acidghost/zotools/internal/sync"
	"github.com/fatih/color"
)

const (
	actCmd    = "act"
	searchCmd = "search"
	syncCmd   = "sync"
)

var (
	bannerC  = color.New(color.FgHiGreen)
	versionC = color.New(color.FgHiCyan)
)

// Set in the Makefile
var version string
//go:embed banner.txt
var banner string

const bannerVersionAlign = "     "

const usageFmt = `Usage: %[1]s ` + OptionsUsage + ` command

Available commands:
  - %[2]s
        download items from Zotero server and update local cache
  - %[3]s
        search for items
  - %[4]s
        execute an action on previous search results

For help on a specific command try: %[1]s command -h

Common options:
`

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), makeBanner()+"\n\n"+usageFmt, os.Args[0],
		syncCmd, searchCmd, actCmd)
	flag.PrintDefaults()
}

func makeBanner() string {
	return fmt.Sprintf("%s\n%s%s", bannerC.Sprint(banner), bannerVersionAlign,
		versionC.Sprint(version))
}

type command interface {
	Run([]string, Config)
}

func main() {
	flagVersion := flag.Bool("V", false, "print version and exit")
	flagConfig := flag.String("config", "", "configuration JSON file (overwrites ZOTOOLS)")
	flagNoColor := flag.Bool("no-color", false, "disable color output")
	flag.Usage = usage
	flag.Parse()

	color.NoColor = *flagNoColor

	// Colorize banner only after `color.NoColor` has been set
	banner := makeBanner()

	if *flagVersion {
		fmt.Println(banner)
		Quit(0)
	}

	// Get remaining arguments that are not part of the root group
	args := os.Args[len(os.Args)-flag.NArg():]
	if len(args) < 1 {
		usage()
		Quit(1)
	}

	if args[0] == "help" {
		usage()
		Quit(0)
	}

	var configPath string

	// Prefer using the one passed as argument over the environment
	if *flagConfig == "" {
		configPath = os.Getenv("ZOTOOLS")
		if configPath == "" {
			Dief("Configuration file is required\n")
		}
	} else {
		configPath = *flagConfig
	}

	config := LoadConfig(configPath)

	var cmd command
	switch args[0] {
	case actCmd:
		cmd = act.New(args[0], banner)
	case searchCmd:
		cmd = search.New(args[0], banner)
	case syncCmd:
		cmd = sync.New(args[0], banner)
	default:
		Dief("Command '%s' not recognized\n", args[0])
	}

	cmd.Run(args[1:], config)
}
