// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type stringBoolPair struct {
	s string
	b bool
}

func tempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "madb_test")
	if err != nil {
		t.Fatalf("could not open a temp file: %v", err)
	}
	f.Close()

	return f.Name()
}

func TestMadbNameSet(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	var got, want map[string]string
	var err error

	// Set a new nickname
	if err = runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"NICKNAME1": "SERIAL1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Set a second nickname
	if err = runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"NICKNAME1": "SERIAL1", "NICKNAME2": "SERIAL2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Override an existing nickname to another
	if err = runMadbNameSet(nil, []string{"SERIAL1", "NN1"}, filename); err != nil {
		t.Fatal(err)
	}

	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"NN1": "SERIAL1", "NICKNAME2": "SERIAL2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Try an existing nickname and see if it fails.
	if err = runMadbNameSet(nil, []string{"SERIAL3", "NN1"}, filename); err == nil {
		t.Fatalf("expected an error but succeeded.")
	}
}

func TestMadbNameUnset(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	// Set up some nicknames first.
	runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename)
	runMadbNameSet(nil, []string{"SERIAL2", "NICKNAME2"}, filename)
	runMadbNameSet(nil, []string{"SERIAL3", "NICKNAME3"}, filename)

	var got, want map[string]string
	var err error

	// Unset by serial number.
	if err = runMadbNameUnset(nil, []string{"SERIAL1"}, filename); err != nil {
		t.Fatal(err)
	}
	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"NICKNAME2": "SERIAL2", "NICKNAME3": "SERIAL3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Unset by nickname.
	if err = runMadbNameUnset(nil, []string{"NICKNAME2"}, filename); err != nil {
		t.Fatal(err)
	}
	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"NICKNAME3": "SERIAL3"}
	if !reflect.DeepEqual(got, want) {
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

	// Run the clear-all command. The file should be empty after running the command.
	runMadbNameClearAll(nil, []string{}, filename)

	// Check if the file is successfully deleted.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("failed to delete file %q", filename)
	}
}

func TestIsValidDeviceSerial(t *testing.T) {
	testCases := []stringBoolPair{
		// The following strings should be accepted
		{"HT4BVWV00023", true},
		{"01023f5e2fd2accf", true},
		{"usb:3-3.4.2", true},
		{"product:volantisg", true},
		{"model:Nexus_9", true},
		{"device:flounder_lte", true},
		{"@1", true},
		{"@2", true},
		// The following strings should not be accepted
		{"have spaces", false},
		{"@abcd", false},
		{"#not_allowed_chars~", false},
	}

	for _, tc := range testCases {
		if got, want := isValidDeviceSerial(tc.s), tc.b; got != want {
			t.Fatalf("unmatched results for serial '%v': got %v, want %v", tc.s, got, want)
		}
	}
}

func TestIsValidNickname(t *testing.T) {
	testCases := []stringBoolPair{
		// The following strings should be accepted
		{"Nexus5X", true},
		{"Nexus9", true},
		{"P1", true},
		{"P2", true},
		{"Tablet", true},
		// The following strings should not be accepted
		{"have spaces", false},
		{"@1", false},
		{"@abcd", false},
		{"#not_allowed_chars~", false},
	}

	for _, tc := range testCases {
		if got, want := isValidNickname(tc.s), tc.b; got != want {
			t.Fatalf("unmatched results for nickname '%v': got %v, want %v", tc.s, got, want)
		}
	}
}

func TestReadNicknameSerialMap(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, "PHONE1 SERIAL1")
	fmt.Fprintln(f, "PHONE2 SERIAL2")
	fmt.Fprintln(f, "PHONE3 SERIAL3")
	f.Close()

	var got map[string]string
	if got, err = readNicknameSerialMap(filename); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"PHONE1": "SERIAL1",
		"PHONE2": "SERIAL2",
		"PHONE3": "SERIAL3",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func TestWriteNicknameSerialMap(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	want := map[string]string{
		"PHONE1": "SERIAL1",
		"PHONE2": "SERIAL2",
		"PHONE3": "SERIAL3",
	}

	if err := writeNicknameSerialMap(want, filename); err != nil {
		t.Fatalf("could not write the map to file: %v", err)
	}

	got, err := readNicknameSerialMap(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
