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

	var cfg *config
	var err error

	// Set a new nickname
	if err = runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.UserIDs, map[string]string{"SERIAL1": "0"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Set a second nickname
	if err = runMadbUserSet(nil, []string{"SERIAL2", "10"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.UserIDs, map[string]string{"SERIAL1": "0", "SERIAL2": "10"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Override an existing nickname to another
	if err = runMadbUserSet(nil, []string{"SERIAL1", "20"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.UserIDs, map[string]string{"SERIAL1": "20", "SERIAL2": "10"}; !reflect.DeepEqual(got, want) {
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

	var cfg *config
	var err error

	// Unset by serial number.
	if err = runMadbUserUnset(nil, []string{"SERIAL1"}, filename); err != nil {
		t.Fatal(err)
	}
	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.UserIDs, map[string]string{"SERIAL2": "0", "SERIAL3": "10"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func ExampleMadbUserList() {
	filename := tempFilename(nil)
	defer os.Remove(filename)

	// Set up some default users first.
	runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL2", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL3", "10"}, filename)

	// Call the list command.
	runMadbUserList(nil, []string{}, filename)

	// Output:
	// +---------+---------+
	// | Serial  | User ID |
	// +---------+---------+
	// | SERIAL1 | 0       |
	// | SERIAL2 | 0       |
	// | SERIAL3 | 10      |
	// +---------+---------+
}

func TestMadbUserClearAll(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some default users first.
	runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL2", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL3", "10"}, filename)

	// Set up some default nicknames. These nicknames should be preserved after
	// running the "user clear-all" command.
	runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename)
	runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename)

	// Run the clear-all command.
	runMadbUserClearAll(nil, []string{}, filename)

	cfg, err := readConfig(filename)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure that the user IDs are all deleted.
	if got, want := cfg.UserIDs, map[string]string{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Make sure that the nicknames are preserved.
	if got, want := cfg.Names, map[string]string{"NICKNAME1": "SERIAL1", "NICKNAME2": "SERIAL2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
