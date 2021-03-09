// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package utils

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

var errorP = color.New(color.FgRed)

var (
	// Quit enables us to get coverage even when calling `os.Exit`
	Quit = os.Exit
)

func Eprintf(format string, args ...interface{}) {
	errorP.Fprintf(os.Stderr, format, args...)
}

func Die(format string, args ...interface{}) {
	Eprintf(format, args...)
	Quit(1)
}

type DummyFS struct{}

func (*DummyFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

const (
	OptionsUsage = "[OPTIONS]"
	topUsage     = "Usage: %s " + OptionsUsage + " %s%s\n\nCommon options:\n"
	bottomUsage  = "\nCommand specific arguments and options:\n"
)

func MakeUsage(flags *flag.FlagSet, cmd, banner, top, bottom string) func() {
	return func() {
		fmt.Fprintf(flag.CommandLine.Output(), banner+"\n\n"+topUsage, os.Args[0], cmd, top)
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), bottomUsage+bottom)
		flags.PrintDefaults()
	}
}

func MakePath(base, key, filename string) string {
	return filepath.Join(base, "storage", key, filename)
}
