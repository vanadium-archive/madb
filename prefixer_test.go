// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPrefixerSingleDevice(t *testing.T) {
	buffer := bytes.Buffer{}
	prefixer := newPrefixer(&buffer, "deviceid01")

	fmt.Fprintln(prefixer, "First line.")
	fmt.Fprintln(prefixer, "Second line.")
	prefixer.Close()

	want := `[deviceid01]	First line.
[deviceid01]	Second line.
`

	if got := buffer.String(); got != want {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func TestPrefixerTwoDevices(t *testing.T) {
	// Two devices case
	buffer := bytes.Buffer{}
	prefixer1 := newPrefixer(&buffer, "deviceid01")
	prefixer2 := newPrefixer(&buffer, "deviceid02")

	fmt.Fprintf(prefixer2, "Second")
	fmt.Fprintln(prefixer1, "First line.")
	fmt.Fprintln(prefixer2, " line.")
	fmt.Fprintln(prefixer1, "Third line.")
	fmt.Fprintln(prefixer2, "Fourth line.")
	prefixer1.Close()
	prefixer2.Close()

	want := `[deviceid01]	First line.
[deviceid02]	Second line.
[deviceid01]	Third line.
[deviceid02]	Fourth line.
`

	if got := buffer.String(); got != want {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func TestPrefixerLastLine(t *testing.T) {
	// For the last line, a newline character should be added automatically.
	buffer := bytes.Buffer{}
	prefixer := newPrefixer(&buffer, "deviceid01")

	fmt.Fprintf(prefixer, "First line.")
	prefixer.Close()

	want := `[deviceid01]	First line.
`

	if got := buffer.String(); got != want {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
