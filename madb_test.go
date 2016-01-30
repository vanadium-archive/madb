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
deviceid01          device usb:3-3.4.3 product:bullhead model:Nexus_5X device:bullhead
emulator-5554       device product:sdk_phone_armv7 model:sdk_phone_armv7 device:generic

`

	got, err := parseDevicesOutput(output, nil)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}

	want := []device{
		device{
			Serial:     "deviceid01",
			Type:       realDevice,
			Qualifiers: []string{"usb:3-3.4.3", "product:bullhead", "model:Nexus_5X", "device:bullhead"},
			Nickname:   "",
		},
		device{
			Serial:     "emulator-5554",
			Type:       emulator,
			Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
			Nickname:   "",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// No devices at all
	output = `List of devices attached

`

	got, err = parseDevicesOutput(output, nil)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}
	if want = []device{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Offline devices should be excluded
	output = `List of devices attached
deviceid01       offline usb:3-3.4.3 product:bullhead model:Nexus_5X device:bullhead
deviceid02       device product:sdk_phone_armv7 model:sdk_phone_armv7 device:generic

`
	got, err = parseDevicesOutput(output, nil)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}

	want = []device{
		device{
			Serial:     "deviceid02",
			Type:       realDevice,
			Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
			Nickname:   "",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// In case some nicknames are defined.
	output = `List of devices attached
deviceid01          device usb:3-3.4.3 product:bullhead model:Nexus_5X device:bullhead
emulator-5554       device product:sdk_phone_armv7 model:sdk_phone_armv7 device:generic

	`

	nsm := map[string]string{
		"MyPhone": "deviceid01",
		"ARMv7":   "model:sdk_phone_armv7",
	}

	got, err = parseDevicesOutput(output, nsm)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}

	want = []device{
		device{
			Serial:     "deviceid01",
			Type:       realDevice,
			Qualifiers: []string{"usb:3-3.4.3", "product:bullhead", "model:Nexus_5X", "device:bullhead"},
			Nickname:   "MyPhone",
		},
		device{
			Serial:     "emulator-5554",
			Type:       emulator,
			Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
			Nickname:   "ARMv7",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}
