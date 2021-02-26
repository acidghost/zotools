// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeUsage(t *testing.T) {
	b := bytes.NewBufferString("")
	fs := flag.NewFlagSet("cmd", flag.ExitOnError)
	fs.SetOutput(b)
	const arg0 = "zotools"
	cmd, banner, top, bottom := "cmd", "banner", "topstuff", "bottomstuff"
	usage := MakeUsage(fs, cmd, banner, top, bottom)

	// Do some trickery to capture `usage` output
	oldCmdLine := flag.CommandLine
	flag.CommandLine = fs
	oldArg0 := os.Args[0]
	os.Args[0] = arg0

	usage()

	// Put everything where it was
	flag.CommandLine = oldCmdLine
	os.Args[0] = oldArg0

	bs := b.String()
	assert.Contains(t, bs, banner)
	assert.Contains(t, bs, "Usage: "+arg0+" "+OptionsUsage+" "+cmd)
	assert.Contains(t, bs, top)
	assert.Contains(t, bs, bottom)
}
