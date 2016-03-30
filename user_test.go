// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"
)

func TestMadbUserSet(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	var got, want map[string]string
	var err error

	// Set a new nickname
	if err = runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readMapFromFile(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"SERIAL1": "0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Set a second nickname
	if err = runMadbUserSet(nil, []string{"SERIAL2", "10"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readMapFromFile(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"SERIAL1": "0", "SERIAL2": "10"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Override an existing nickname to another
	if err = runMadbUserSet(nil, []string{"SERIAL1", "20"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readMapFromFile(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"SERIAL1": "20", "SERIAL2": "10"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Try some invalid ids and see if they fail.
	invalidIds := []string{"-1", "NAME"}
	for _, id := range invalidIds {
		if err = runMadbUserSet(nil, []string{"SERIAL", id}, filename); err == nil {
			t.Fatalf("expected an error but succeeded.")
		}
	}
}

func TestMadbUserUnset(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some default users first.
	runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL2", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL3", "10"}, filename)

	var got, want map[string]string
	var err error

	// Unset by serial number.
	if err = runMadbUserUnset(nil, []string{"SERIAL1"}, filename); err != nil {
		t.Fatal(err)
	}
	if got, err = readMapFromFile(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"SERIAL2": "0", "SERIAL3": "10"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func TestMadbUserClearAll(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some default users first.
	runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL2", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL3", "10"}, filename)

	// Run the clear-all command. The file should be empty after running the command.
	runMadbUserClearAll(nil, []string{}, filename)

	// Check if the file is successfully deleted.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("failed to delete file %q", filename)
	}
}
