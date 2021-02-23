// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU GPL License version 3.

package main

import (
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

var bannerColor = color.New(color.FgHiGreen)
var banner = `                  __                 ___
                 /\ \__             /\_ \            
     ____     ___\ \ ,_\   ___    __\//\ \     ____  
    /\_ ,` + "`" + `\  / __` + "`" + `\ \ \/  / __` + "`" + `\ / __` + "`" + `\ \ \   /',__\ 
    \/_/  /_/\ \L\ \ \ \_/\ \L\ /\ \L\ \_\ \_/\__, ` + "`" + `\
      /\____\ \____/\ \__\ \____\ \____/\____\/\____/
      \/____/\/___/  \/__/\/___/ \/___/\/____/\/___/ 
`

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
	b := bannerColor.Sprint(banner)
	fmt.Fprintf(flag.CommandLine.Output(), b+"\n\n"+usageFmt, os.Args[0],
		syncCmd, searchCmd, actCmd)
	flag.PrintDefaults()
}

type command interface {
	Run([]string, Config)
}

func main() {
	flagConfig := flag.String("config", "", "configuration JSON file (overwrites ZOTOOLS)")
	flagNoColor := flag.Bool("no-color", false, "disable color output")
	flag.Usage = usage
	flag.Parse()

	color.NoColor = *flagNoColor

	// Get remaining arguments that are not part of the root group
	args := os.Args[len(os.Args)-flag.NArg():]
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	if args[0] == "help" {
		usage()
		os.Exit(0)
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

	banner := bannerColor.Sprint(banner)
	var cmd command
	switch args[0] {
	case actCmd:
		cmd = act.New(args[0], banner)
	case searchCmd:
		cmd = search.New(args[0], banner)
	case syncCmd:
		cmd = sync.New(args[0], banner)
	default:
		Dief("Command not recognized\n")
	}

	cmd.Run(args[1:], config)
}
