// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"

	"v.io/x/lib/cmdline"
)

type testCase struct {
	runnerFunc  func(*cmdline.Env, []string, string) error
	input       []string
	want        map[string][]string
	errExpected bool
}

type testSequence []testCase

func runGroupTests(t *testing.T, tests []testSequence) {
	for i, testSuite := range tests {
		filename := tempFilename(t)
		defer os.Remove(filename)

		for j, test := range testSuite {
			err := test.runnerFunc(cmdline.EnvFromOS(), test.input, filename)
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

func TestMadbGroupAdd(t *testing.T) {
	tests := []testSequence{
		{
			{
				runMadbGroupAdd,
				[]string{},
				map[string][]string{},
				true,
			},
		},
		{
			{
				runMadbGroupAdd,
				[]string{"GROUP1"},
				map[string][]string{},
				true,
			},
		},
		{
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME2"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2"}},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL2", "NICKNAME3", "NICKNAME1"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"}},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"GROUP2", "SERIAL1", "SERIAL4", "NICKNAME1"},
				map[string][]string{
					"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
					"GROUP2": []string{"SERIAL1", "SERIAL4", "NICKNAME1"},
				},
				false,
			},
			{
				runMadbGroupAdd,
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
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME1"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1"}},
				false,
			},
		},
		{
			// Invalid mamber name
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "#INVALID_NAME#"},
				map[string][]string{},
				true,
			},
			// Invalid member index
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "@ABC"},
				map[string][]string{},
				true,
			},
		},
	}

	runGroupTests(t, tests)
}

func TestMadbGroupAddNameConflict(t *testing.T) {
	tests := []testSequence{
		{
			{
				runMadbNameSet,
				[]string{"SERIAL1", "NICKNAME1"},
				map[string][]string{},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"NICKNAME1", "SERIAL2"},
				map[string][]string{},
				true,
			},
		},
	}

	runGroupTests(t, tests)
}

func TestMadbGroupRemove(t *testing.T) {
	tests := []testSequence{
		{
			{
				runMadbGroupRemove,
				[]string{},
				map[string][]string{},
				true,
			},
		},
		{
			{
				runMadbGroupRemove,
				[]string{"GROUP1"},
				map[string][]string{},
				true,
			},
		},
		{
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"}},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"GROUP2", "SERIAL3", "NICKNAME4"},
				map[string][]string{
					"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
				},
				false,
			},
			{
				runMadbGroupRemove,
				[]string{"GROUP1", "SERIAL1", "NICKNAME3"},
				map[string][]string{
					"GROUP1": []string{"NICKNAME1", "NICKNAME2", "SERIAL2"},
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
				},
				false,
			},
			{
				runMadbGroupRemove,
				[]string{"GROUP2", "SERIAL3"},
				map[string][]string{
					"GROUP1": []string{"NICKNAME1", "NICKNAME2", "SERIAL2"},
					"GROUP2": []string{"NICKNAME4"},
				},
				false,
			},
			{
				runMadbGroupRemove,
				[]string{"GROUP3", "SERIAL3"},
				map[string][]string{
					"GROUP1": []string{"NICKNAME1", "NICKNAME2", "SERIAL2"},
					"GROUP2": []string{"NICKNAME4"},
				},
				true,
			},
			{
				runMadbGroupRemove,
				[]string{"GROUP2", "NICKNAME4"},
				map[string][]string{
					"GROUP1": []string{"NICKNAME1", "NICKNAME2", "SERIAL2"},
				},
				false,
			},
			{
				runMadbGroupRemove,
				[]string{"GROUP1", "NICKNAME2", "SERIAL2", "NICKNAME1"},
				map[string][]string{},
				false,
			},
		},
	}

	runGroupTests(t, tests)
}

func TestMadbGroupRename(t *testing.T) {
	tests := []testSequence{
		{
			{
				runMadbGroupRename,
				[]string{},
				map[string][]string{},
				true,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP1"},
				map[string][]string{},
				true,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP1", "GROUP2"},
				map[string][]string{},
				true,
			},
		},
		{
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
				map[string][]string{"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"}},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"GROUP2", "SERIAL3", "NICKNAME4"},
				map[string][]string{
					"GROUP1": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
				},
				false,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP1", "GROUP3"},
				map[string][]string{
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
					"GROUP3": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
				},
				false,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP2", "_!@#"},
				map[string][]string{
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
					"GROUP3": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
				},
				true,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP2", "GROUP3"},
				map[string][]string{
					"GROUP2": []string{"SERIAL3", "NICKNAME4"},
					"GROUP3": []string{"SERIAL1", "NICKNAME1", "NICKNAME2", "SERIAL2", "NICKNAME3"},
				},
				true,
			},
		},
	}

	runGroupTests(t, tests)
}

func TestMadbGroupRenameNameConflict(t *testing.T) {
	tests := []testSequence{
		{
			{
				runMadbNameSet,
				[]string{"SERIAL1", "NICKNAME1"},
				map[string][]string{},
				false,
			},
			{
				runMadbGroupAdd,
				[]string{"GROUP1", "SERIAL1"},
				map[string][]string{"GROUP1": []string{"SERIAL1"}},
				false,
			},
			{
				runMadbGroupRename,
				[]string{"GROUP1", "NICKNAME1"},
				map[string][]string{"GROUP1": []string{"SERIAL1"}},
				true,
			},
		},
	}

	runGroupTests(t, tests)
}
