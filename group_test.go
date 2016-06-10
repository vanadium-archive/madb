// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"
)

func TestMadbGroupAdd(t *testing.T) {
	tests := [][]struct {
		input       []string
		want        map[string][]string
		errExpected bool
	}{
		{
			{
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME2"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2"}},
				false,
			},
			{
				[]string{"GROUP1", "SERIAL2", "NICKNAME3", "NICKNAME1"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"}},
				false,
			},
			{
				[]string{"GROUP2", "SERIAL1", "SERIAL4", "NICKNAME1"},
				map[string][]string{
					"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
					"GROUP2": []string{"SERIAL1", "SERIAL4", "NICKNAME1"},
				},
				false,
			},
			{
				[]string{"GROUP2", "GROUP1"},
				map[string][]string{
					"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
					"GROUP2": []string{"SERIAL1", "SERIAL4", "NICKNAME1", "GROUP1"},
				},
				false,
			},
		},
		{
			// Duplicate members should be removed even within the arguments.
			{
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME1"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1"}},
				false,
			},
		},
		{
			// Invalid mamber name
			{
				[]string{"GROUP1", "#INVALID_NAME#"},
				map[string][]string{},
				true,
			},
			// Invalid member index
			{
				[]string{"GROUP1", "@ABC"},
				map[string][]string{},
				true,
			},
		},
	}

	for i, testSuite := range tests {
		filename := tempFilename(t)
		defer os.Remove(filename)

		for j, test := range testSuite {
			err := runMadbGroupAdd(nil, test.input, filename)
			if test.errExpected != (err != nil) {
				t.Fatalf("error expected for tests[%v][%v]: %v, got: %v", i, j, test.errExpected, err)
			}

			cfg, err := readConfig(filename)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := cfg.Groups, test.want; !reflect.DeepEqual(got, want) {
				t.Fatalf("unmatched results for tests[%v][%v]: got %v, want %v", i, j, got, want)
			}
		}
	}
}

func TestMadbGroupAddNameConflict(t *testing.T) {
	filename := tempFilename(t)
	defer os.Remove(filename)

	err := runMadbNameSet(nil, []string{"SERIAL1", "NICKNAME1"}, filename)
	if err != nil {
		t.Fatal(err)
	}

	err = runMadbGroupAdd(nil, []string{"NICKNAME1", "SERIAL2"}, filename)
	if err == nil {
		t.Fatalf("error expected but got: %v", err)
	}
}
