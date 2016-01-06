// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"
)

func TestParseDevicesOutput(t *testing.T) {
	var output string

	// Normal case
	output = `List of devices attached
deviceid01       device usb:3-3.4.3 product:bullhead model:Nexus_5X device:bullhead
deviceid02       device product:sdk_phone_armv7 model:sdk_phone_armv7 device:generic

`

	result, err := parseDevicesOutput(output)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}
	if got, want := result, []string{"deviceid01", "deviceid02"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// No devices at all
	output = `List of devices attached

`

	result, err = parseDevicesOutput(output)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}
	if got, want := result, []string{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Offline devices should be excluded
	output = `List of devices attached
deviceid01       offline usb:3-3.4.3 product:bullhead model:Nexus_5X device:bullhead
deviceid02       device product:sdk_phone_armv7 model:sdk_phone_armv7 device:generic

`
	result, err = parseDevicesOutput(output)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}
	if got, want := result, []string{"deviceid02"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
