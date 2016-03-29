// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
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

func TestParseScreenDensity(t *testing.T) {
	tests := []struct {
		outputFile string
		want       int
	}{
		{filepath.Join("testdata", "getprop1.txt"), 420},
		{filepath.Join("testdata", "getprop2.txt"), 320},
	}

	for i, test := range tests {
		f, err := os.Open(test.outputFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		got, err := parseScreenDensity(string(bytes))
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, test.want)
		}
	}
}

// TODO(youngseokyoon): restructure the Gradle test projects so that they share the binary data.
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
	if got, want := len(props.VariantOutputs), 3; got != want {
		t.Fatalf("unexpected variant output length: got %v, want %v", got, want)
	}

	outputX86 := &props.VariantOutputs[0]
	outputArmv7a := &props.VariantOutputs[1]
	outputMips := &props.VariantOutputs[2]

	// Now create fake device properties and test whether the computeBestOutput returns the correct
	// variantOutput.
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

func TestComputeBestOutputForDensitySplits(t *testing.T) {
	// Read the variant properties from the "testAndroidDensitySplit" project.
	key := variantKey{
		Dir:     filepath.Join("testdata", "projects", "testAndroidDensitySplit"),
		Module:  "app",
		Variant: "Debug",
	}

	props, err := extractPropertiesFromGradle(key)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure that it has three outputs for each density.
	if got, want := len(props.VariantOutputs), 3; got != want {
		t.Fatalf("unexpected variant output length: got %v, want %v", got, want)
	}

	outputUniversal := &props.VariantOutputs[0]
	outputLdpi := &props.VariantOutputs[1]
	outputMdpi := &props.VariantOutputs[2]

	// Now create fake device properties and test whether the computeBestOutput returns the correct
	// variantOutput.
	deviceAbis := []string{"armeabi-v7a"}
	tests := []struct {
		deviceDensity int
		want          *variantOutput
	}{
		{120, outputLdpi},
		{160, outputMdpi},
		{420, outputUniversal},
	}

	for i, test := range tests {
		if got := computeBestOutput(props.VariantOutputs, props.AbiFilters, test.deviceDensity, deviceAbis); got != test.want {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, test.want)
		}
	}
}
