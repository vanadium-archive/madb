// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"
)

func TestMadbNameSet(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	var cfg *config
	var err error

	// Set a new nickname
	if err = runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Names, map[string]string{"NICKNAME1": "SERIAL1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Set a second nickname
	if err = runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Names, map[string]string{"NICKNAME1": "SERIAL1", "NICKNAME2": "SERIAL2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Override an existing nickname to another
	if err = runMadbNameSet(nil, []string{"SERIAL1", "NN1"}, filename); err != nil {
		t.Fatal(err)
	}

	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Names, map[string]string{"NN1": "SERIAL1", "NICKNAME2": "SERIAL2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Try an existing nickname and see if it fails.
	if err = runMadbNameSet(nil, []string{"SERIAL3", "NN1"}, filename); err == nil {
		t.Fatalf("expected an error but succeeded.")
	}

	// Try an existing group name and see if it fails.
	err = runMadbGroupAdd(nil, []string{"GROUP1", "SERIAL1"}, filename)
	if err != nil {
		t.Fatal(err)
	}

	err = runMadbNameSet(nil, []string{"SERIAL4", "GROUP1"}, filename)
	if err == nil {
		t.Fatalf("error expected but got: %v", err)
	}
}

func TestMadbNameUnset(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some nicknames first.
	runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename)
	runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename)
	runMadbNameSet(nil, []string{"SERIAL3", "NICKNAME3"}, filename)

	var cfg *config
	var err error

	// Unset by serial number.
	if err = runMadbNameUnset(nil, []string{"SERIAL1"}, filename); err != nil {
		t.Fatal(err)
	}
	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Names, map[string]string{"NICKNAME2": "SERIAL2", "NICKNAME3": "SERIAL3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Unset by nickname.
	if err = runMadbNameUnset(nil, []string{"NICKNAME2"}, filename); err != nil {
		t.Fatal(err)
	}
	if cfg, err = readConfig(filename); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Names, map[string]string{"NICKNAME3": "SERIAL3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// When the input is not found anywhere.
	if err = runMadbNameUnset(nil, []string{"UnrecognizedName"}, filename); err == nil {
		t.Fatalf("expected an error but succeeded.")
	}
}

func TestMadbNameList(t *testing.T) {
	// TODO(youngseokyoon): add some tests for the list command.
}

func TestMadbNameClearAll(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some nicknames first.
	runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename)
	runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename)
	runMadbNameSet(nil, []string{"SERIAL3", "NICKNAME3"}, filename)

	// Set up some default users. These users should be preserved after running
	// the "name clear-all" command.
	runMadbUserSet(nil, []string{"SERIAL1", "0"}, filename)
	runMadbUserSet(nil, []string{"SERIAL2", "10"}, filename)

	// Run the clear-all command. The file should be empty after running the command.
	runMadbNameClearAll(nil, []string{}, filename)

	cfg, err := readConfig(filename)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure that the names are all deleted.
	if got, want := cfg.Names, map[string]string{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Make sure that the default user IDs are preserved.
	if got, want := cfg.UserIDs, map[string]string{"SERIAL1": "0", "SERIAL2": "10"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
