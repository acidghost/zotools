package common

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

var errorP = color.New(color.FgRed)

func Eprintf(format string, args ...interface{}) {
	errorP.Fprintf(os.Stderr, format, args...)
}

func Dief(format string, args ...interface{}) {
	Eprintf(format, args...)
	os.Exit(1)
}

const (
	OptionsUsage = "[OPTIONS]"
	topUsage     = "Usage: %s " + OptionsUsage + " %s%s\n\nCommon options:\n"
	bottomUsage  = "\nCommand specific arguments and options:\n"
)

func MakeUsage(fs *flag.FlagSet, cmd, banner, top, bottom string) func() {
	return func() {
		fmt.Fprintf(flag.CommandLine.Output(), banner+"\n\n"+topUsage, os.Args[0], cmd, top)
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), bottomUsage+bottom)
		fs.PrintDefaults()
	}
}

func MakePath(base, key, filename string) string {
	return filepath.Join(base, "storage", key, filename)
}
