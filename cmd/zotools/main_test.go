// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

// +build coverage

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/acidghost/zotools/internal/utils"
)

var exitCode int

func TestMain(m *testing.M) {
	_ = m.Run()
	os.Exit(exitCode)
}

func TestMainWrap(_ *testing.T) {
	var args []string

	// Collect arguments for `main`
	for _, arg := range os.Args {
		switch {
		case strings.HasPrefix(arg, "-test"):
			// Skip flags for the test runner
		case strings.HasPrefix(arg, "COVERAGE"):
			// Dummy argument to make the test runner stop parsing flags
		default:
			args = append(args, arg)
		}
	}

	os.Args = args

	// Replace `Quit` so we don't exit prematurely and store coverage appropriately
	ch := make(chan int)
	oldQuit := utils.Quit
	defer func() { utils.Quit = oldQuit }()
	utils.Quit = func(code int) {
		ch <- code
		// Lock the goroutine running `main`, otherwise returning will cause havoc
		ch <- code
	}

	go func() {
		main()
		// If main does not call `Quit`, close the channel
		close(ch)
	}()

	// Wait for `main` to call `Quit` or return
	if c, ok := <-ch; ok {
		exitCode = c
	}

	null, _ := os.Open(os.DevNull)
	os.Stdout = null
	os.Stderr = null
}
