// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseSupportedAbis(t *testing.T) {
	tests := []struct {
		output string
		want   []string
	}{
		{
			`config: en-rUS-ldltr-sw411dp-w411dp-h659dp-normal-notlong-notround-port-notnight-420dpi-finger-keysexposed-nokeys-navhidden-nonav-v23
abi: arm64-v8a,armeabi-v7a,armeabi
`,
			[]string{"arm64-v8a", "armeabi-v7a", "armeabi"},
		},
	}

	for i, test := range tests {
		got, err := parseSupportedAbis(test.output)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, test.want)
		}
	}
}

// TODO(youngseokyoon): add tests for the density splits too.
func TestComputeBestOutputForAbiSplits(t *testing.T) {
	// Read the variant properties from the "testAndroidAbiSplit" project.
	key := variantKey{
		Dir:     filepath.Join("testdata", "projects", "testAndroidAbiSplit"),
		Module:  "app",
		Variant: "Debug",
	}

	props, err := extractPropertiesFromGradle(key)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure that it has three outputs for each Abi.
	if len(props.VariantOutputs) != 3 {
		t.Fatalf("The number of extracted variant outputs does not match the expected value.")
	}

	outputX86 := &props.VariantOutputs[0]
	outputArmv7a := &props.VariantOutputs[1]
	outputMips := &props.VariantOutputs[2]

	// Now create fake device properties and test whether the computeBestOutput returns the correct variantOutput.
	deviceDensity := 240
	tests := []struct {
		deviceAbis []string
		want       *variantOutput
	}{
		{[]string{"x86"}, outputX86},
		{[]string{"armeabi-v7a"}, outputArmv7a},
		{[]string{"mips"}, outputMips},
		{[]string{"x86", "armeabi-v7a"}, outputX86},
		{[]string{"x86_64"}, nil},
	}

	for i, test := range tests {
		if got := computeBestOutput(props.VariantOutputs, props.AbiFilters, deviceDensity, test.deviceAbis); got != test.want {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, test.want)
		}
	}
}
