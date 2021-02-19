package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	bold  = color.New(color.Bold).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
	blue  = color.New(color.FgBlue).SprintFunc()
)

const Banner = `                  __                 ___             
                 /\ \__             /\_ \            
     ____     ___\ \ ,_\   ___    __\//\ \     ____  
    /\_ ,` + "`" + `\  / __` + "`" + `\ \ \/  / __` + "`" + `\ / __` + "`" + `\ \ \   /',__\ 
    \/_/  /_/\ \L\ \ \ \_/\ \L\ /\ \L\ \_\ \_/\__, ` + "`" + `\
      /\____\ \____/\ \__\ \____\ \____/\____\/\____/
      \/____/\/___/  \/__/\/___/ \/___/\/____/\/___/ 
`

const UsageFmt = `Usage: %s [OPTIONS] command

Available commands:
  - sync
        download items from Zotero server and update local cache
  - search
        search for items

Common options:
`

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), green(Banner)+"\n\n"+UsageFmt, os.Args[0])
	flag.PrintDefaults()
}

type Config struct {
	Key    string
	Zotero string
	Cache  string
}

type Command interface {
	Run([]string, Config)
	Name() string
}

func main() {
	flagConfig := flag.String("config", "", "configuration JSON file")
	flagNoColor := flag.Bool("no-color", false, "disable color output")
	flag.Usage = Usage
	flag.Parse()

	color.NoColor = *flagNoColor

	cmds := []Command{NewSyncCommand(), NewSearchCommand()}

	// Get remaining arguments that are not part of the root group
	args := os.Args[len(os.Args)-flag.NArg():]
	if len(args) < 1 {
		Usage()
		os.Exit(1)
	}

	if args[0] == "help" {
		Usage()
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

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		Dief("Failed to read config file:\n - %v\n", err)
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		Dief("Failed to parse config JSON from %s: %v\n", *flagConfig, err)
	}

	for _, cmd := range cmds {
		if cmd.Name() == args[0] {
			cmd.Run(args[1:], config)
			os.Exit(0)
		}
	}

	Dief("Command not recognized\n")
}
