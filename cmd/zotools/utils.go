package main

import (
	"os"

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
