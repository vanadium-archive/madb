// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"v.io/x/lib/gosh"
)

func TestMain(m *testing.M) {
	gosh.InitMain()
	os.Exit(m.Run())
}

func tempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "madb_test")
	if err != nil {
		t.Fatalf("could not open a temp file: %v", err)
	}
	f.Close()

	return f.Name()
}

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
			Index:      1,
			UserID:     "",
		},
		device{
			Serial:     "emulator-5554",
			Type:       emulator,
			Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
			Nickname:   "",
			Index:      2,
			UserID:     "",
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
			Index:      2,
			UserID:     "",
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

	cfg := &config{
		Names: map[string]string{
			"MyPhone": "deviceid01",
			"ARMv7":   "model:sdk_phone_armv7",
		},
		UserIDs: map[string]string{
			"deviceid01": "10",
		},
	}

	got, err = parseDevicesOutput(output, cfg)
	if err != nil {
		t.Fatalf("failed to parse the output: %v", err)
	}

	want = []device{
		device{
			Serial:     "deviceid01",
			Type:       realDevice,
			Qualifiers: []string{"usb:3-3.4.3", "product:bullhead", "model:Nexus_5X", "device:bullhead"},
			Nickname:   "MyPhone",
			Index:      1,
			UserID:     "10",
		},
		device{
			Serial:     "emulator-5554",
			Type:       emulator,
			Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
			Nickname:   "ARMv7",
			Index:      2,
			UserID:     "",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

func TestGetSpecifiedDevices(t *testing.T) {
	// First, define some devices (three real devices, and two emulators).
	d1 := device{
		Serial:     "deviceid01",
		Type:       realDevice,
		Qualifiers: []string{"usb:3-3.4.3", "product:bullhead", "model:Nexus_5X", "device:bullhead"},
		Nickname:   "MyPhone",
		Index:      1,
		UserID:     "",
	}

	d2 := device{
		Serial:     "deviceid02",
		Type:       realDevice,
		Qualifiers: []string{"usb:3-3.4.1", "product:volantisg", "model:Nexus_9", "device:flounder_lte"},
		Nickname:   "",
		Index:      2,
		UserID:     "",
	}

	e1 := device{
		Serial:     "emulator-5554",
		Type:       emulator,
		Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
		Nickname:   "ARMv7",
		Index:      3,
		UserID:     "",
	}

	d3 := device{
		Serial:     "deviceid03",
		Type:       realDevice,
		Qualifiers: []string{"usb:3-3.3", "product:bullhead", "model:Nexus_5X", "device:bullhead"},
		Nickname:   "SecondPhone",
		Index:      4,
		UserID:     "",
	}

	e2 := device{
		Serial:     "emulator-5555",
		Type:       emulator,
		Qualifiers: []string{"product:sdk_phone_armv7", "model:sdk_phone_armv7", "device:generic"},
		Nickname:   "",
		Index:      5,
		UserID:     "",
	}

	allDevices := []device{d1, d2, e1, d3, e2}

	type deviceFlags struct {
		allDevices   bool
		allEmulators bool
		devices      string
	}

	testCases := []struct {
		flags deviceFlags
		want  []device
	}{
		{deviceFlags{false, false, ""}, allDevices},                                         // Nothing is specified
		{deviceFlags{true, true, ""}, allDevices},                                           // Both -d and -e are specified
		{deviceFlags{true, false, ""}, []device{d1, d2, d3}},                                // Only -d is specified
		{deviceFlags{false, true, ""}, []device{e1, e2}},                                    // Only -e is specified
		{deviceFlags{false, false, "device:bullhead"}, []device{d1, d3}},                    // Device qualifier
		{deviceFlags{false, false, "ARMv7,SecondPhone"}, []device{e1, d3}},                  // Nicknames
		{deviceFlags{false, false, "@2,@4"}, []device{d2, d3}},                              // Device Indices
		{deviceFlags{false, false, "NormalGroup"}, []device{d1, d2, e1}},                    // Normal group
		{deviceFlags{false, false, "SelfRefGroup"}, []device{d2}},                           // Self referencing group
		{deviceFlags{false, false, "CyclicGroup1"}, []device{d1, d2, d3}},                   // Cyclic group inclusion
		{deviceFlags{true, false, "ARMv7"}, []device{d1, d2, e1, d3}},                       // Combinations
		{deviceFlags{false, true, "model:Nexus_9"}, []device{d2, e1, e2}},                   // Combinations
		{deviceFlags{false, false, "@1,SecondPhone"}, []device{d1, d3}},                     // Combinations
		{deviceFlags{false, false, "SecondPhone,NormalGroup,@1"}, []device{d1, d2, e1, d3}}, // Combinations
	}

	cfg := &config{
		Groups: map[string][]string{
			"NormalGroup":  []string{"deviceid01", "deviceid02", "@3"},
			"SelfRefGroup": []string{"deviceid02", "SelfRefGroup"},
			"CyclicGroup1": []string{"CyclicGroup2", "@1"},
			"CyclicGroup2": []string{"@2", "CyclicGroup3"},
			"CyclicGroup3": []string{"deviceid03", "CyclicGroup1"},
		},
	}

	for i, testCase := range testCases {
		allDevicesFlag = testCase.flags.allDevices
		allEmulatorsFlag = testCase.flags.allEmulators
		devicesFlag = testCase.flags.devices

		got, err := filterSpecifiedDevices(allDevices, cfg)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("unmatched results for testCases[%v]: got %v, want %v", i, got, testCase.want)
		}
	}
}

func TestIsFlutterProject(t *testing.T) {
	testCases := []struct {
		projectDir string
		want       bool
	}{
		{"testMultiPlatform", false},
		{"testMultiPlatform/android", false},
		{"testMultiPlatform/android/app", false},
		{"testMultiPlatform/flutter", true},
	}

	for i, testCase := range testCases {
		dir := filepath.Join("testdata", "projects", testCase.projectDir)
		if got := isFlutterProject(dir); got != testCase.want {
			t.Fatalf("unmatched results for testCases[%v]: got %v, want %v", i, got, testCase.want)
		}
	}
}

func TestIsGradleProject(t *testing.T) {
	testCases := []struct {
		projectDir string
		want       bool
	}{
		{"testMultiPlatform", false},
		{"testMultiPlatform/android", true},
		{"testMultiPlatform/android/app", true},
		{"testMultiPlatform/flutter", false},
	}

	for i, testCase := range testCases {
		dir := filepath.Join("testdata", "projects", testCase.projectDir)
		if got := isGradleProject(dir); got != testCase.want {
			t.Fatalf("unmatched results for testCases[%v]: got %v, want %v", i, got, testCase.want)
		}
	}
}

func TestExtractPropertiesFromGradle(t *testing.T) {
	tests := []struct {
		key  variantKey
		want variantProperties
	}{
		{
			variantKey{"testMultiPlatform/android", "", ""},
			variantProperties{AppID: "io.v.testProjectId", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testMultiPlatform/android", "app", "debug"},
			variantProperties{AppID: "io.v.testProjectId", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testMultiPlatform/android/app", "", ""},
			variantProperties{AppID: "io.v.testProjectId", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testAndroidMultiFlavor", "", ""},
			variantProperties{AppID: "io.v.testProjectId.lite.debug", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testAndroidMultiFlavor", "app", "liteDebug"},
			variantProperties{AppID: "io.v.testProjectId.lite.debug", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testAndroidMultiFlavor/app", "", "proRelease"},
			variantProperties{AppID: "io.v.testProjectId.pro", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
		{
			variantKey{"testApplicationIdFallback", "", ""},
			variantProperties{AppID: "io.v.testProjectPackage", Activity: "io.v.testProjectPackage.LauncherActivity"},
		},
	}

	for i, test := range tests {
		test.key.Dir = filepath.Join("testdata", "projects", test.key.Dir)
		got, err := extractPropertiesFromGradle(test.key)
		if err != nil {
			t.Fatalf("error occurred while extracting properties for testCases[%v]: %v", i, err)
		}

		if got.AppID != test.want.AppID || got.Activity != test.want.Activity {
			t.Fatalf("unmatched results for testCases[%v]: got %v, want %v", i, got, test.want)
		}
	}
}

func TestGetProjectProperties(t *testing.T) {
	cacheFile := tempFilename(t)
	defer os.Remove(cacheFile)

	called := false

	// See if it runs the extractor for the first time.
	extractor := func(key variantKey) (variantProperties, error) {
		called = true
		return variantProperties{AppID: "testAppID", Activity: "Activity"}, nil
	}

	want := variantProperties{AppID: "testAppID", Activity: "Activity"}
	got, err := getProjectProperties(extractor, variantKey{"testDir", "mod", "var"}, false, cacheFile)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if !called {
		t.Fatalf("extractor was not called when expected to be called.")
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// The second run should not invoke the extractor.
	called = false
	got, err = getProjectProperties(extractor, variantKey{"testDir", "mod", "var"}, false, cacheFile)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if called {
		t.Fatalf("extracted was called when not expected.")
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}

	// Run with clear cache flag.
	called = false
	got, err = getProjectProperties(extractor, variantKey{"testDir", "mod", "var"}, true, cacheFile)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if !called {
		t.Fatalf("extractor was not called when expected to be called.")
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unmatched results: got %v, want %v", got, want)
	}
}

// TestEmbeddedGradleScript tests whether the gradle script defined in embedded_gradle.go matches
// the madb_init.gradle file.
func TestEmbeddedGradleScript(t *testing.T) {
	f, err := os.Open("madb_init.gradle")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != gradleInitScript {
		t.Fatalf(`The embedded Gradle script is out of date. Please run "jiri go generate" to regenerate the embedded script.`)
	}
}

// TestExpandKeywords tests the "expandKeywords" function, which is used for expanding pre-defined
// keywords such as "{{serial}}" and "{{name}}".
func TestExpandKeywords(t *testing.T) {
	// Sample devices.
	d1 := device{
		Serial:     "0123456789",
		Type:       realDevice,
		Qualifiers: nil,
		Nickname:   "Alice",
		Index:      1,
		UserID:     "10",
	}

	d2 := device{
		Serial:     "emulator-1234",
		Type:       emulator,
		Qualifiers: nil,
		Nickname:   "",
		Index:      2,
		UserID:     "",
	}

	tests := []struct {
		arg  string
		d    device
		want string
	}{
		{"{{name}}.txt", d1, "Alice.txt"},
		{"{{serial}}.txt", d1, "0123456789.txt"},
		{"{{index}}.txt", d1, "1.txt"},
		{"Hello, {{name}}!", d2, "Hello, emulator-1234!"},
		{"Hello, {{serial}}!", d2, "Hello, emulator-1234!"},
		{"{{index}}.txt", d2, "2.txt"},
		{"{{name}}-{{serial}}.txt", d1, "Alice-0123456789.txt"},
	}

	for i, test := range tests {
		if got := expandKeywords(test.arg, test.d); got != test.want {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, test.want)
		}
	}
}

var helloFunc = gosh.RegisterFunc("helloFunc", func() {
	fmt.Println("Hello, World!")
})

func TestOutputPrefix(t *testing.T) {
	// Sample device.
	d1 := device{
		Serial:     "deviceid01",
		Type:       realDevice,
		Qualifiers: nil,
		Nickname:   "Alice",
		Index:      1,
		UserID:     "",
	}

	d2 := device{
		Serial:     "deviceid02",
		Type:       realDevice,
		Qualifiers: nil,
		Nickname:   "Bob",
		Index:      2,
		UserID:     "10",
	}

	d3 := device{
		Serial:     "deviceid03",
		Type:       realDevice,
		Qualifiers: nil,
		Nickname:   "",
		Index:      3,
		UserID:     "",
	}

	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	tests := []struct {
		prefixType string
		d          device
		want       string
	}{
		{"name", d1, "[Alice]\tHello, World!\n"},
		{"name", d2, "[Bob:10]\tHello, World!\n"},
		{"name", d3, "[deviceid03]\tHello, World!\n"},
		{"serial", d1, "[deviceid01]\tHello, World!\n"},
		{"serial", d2, "[deviceid02:10]\tHello, World!\n"},
		{"serial", d3, "[deviceid03]\tHello, World!\n"},
		{"none", d1, "Hello, World!\n"},
		{"none", d2, "Hello, World!\n"},
		{"none", d3, "Hello, World!\n"},
	}

	for i, test := range tests {
		var b1, b2 bytes.Buffer

		helloCmd := sh.FuncCmd(helloFunc)
		prefixFlag = test.prefixType
		if err := runGoshCommandForDeviceWithWriters(helloCmd, test.d, true, &b1, &b2); err != nil {
			t.Fatalf("error occurred while running gosh command: %v", err)
		}

		if got, want := b1.String(), test.want; got != want {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, want)
		}
		if b2.String() != "" {
			t.Fatalf("unexpected output to stderr for tests[%v]: %v", i, b2.String())
		}
	}
}

func TestIsValidSerial(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
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

	for _, test := range tests {
		if got := isValidSerial(test.input); got != test.want {
			t.Fatalf("unmatched results for serial '%v': got %v, want %v", test.input, got, test.want)
		}
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
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

	for _, test := range tests {
		if got := isValidName(test.input); got != test.want {
			t.Fatalf("unmatched results for nickname '%v': got %v, want %v", test.input, got, test.want)
		}
	}
}

func TestConfigMigration(t *testing.T) {
	tests := []struct {
		configDir string
		fileMap   map[string]string
		want      config
	}{
		{
			"testdata/configs/newFormat",
			map[string]string{"config": "config"},
			config{
				Version: version,
				Names:   map[string]string{"nickname01": "serial01", "nickname02": "serial02"},
				Groups:  map[string][]string{},
				UserIDs: map[string]string{"serial01": "10"},
			},
		},
		{
			"testdata/configs/oldFormatBoth",
			map[string]string{"": "config", "nicknames": "nicknames.bak", "users": "users.bak"},
			config{
				Version: version,
				Names:   map[string]string{"nickname01": "serial01", "nickname02": "serial02"},
				Groups:  map[string][]string{},
				UserIDs: map[string]string{"serial01": "10"},
			},
		},
		{
			"testdata/configs/oldFormatNicknamesOnly",
			map[string]string{"": "config", "nicknames": "nicknames.bak"},
			config{
				Version: version,
				Names:   map[string]string{"nickname01": "serial01", "nickname02": "serial02"},
				Groups:  map[string][]string{},
				UserIDs: map[string]string{},
			},
		},
		{
			"testdata/configs/oldFormatUsersOnly",
			map[string]string{"": "config", "users": "users.bak"},
			config{
				Version: version,
				Names:   map[string]string{},
				Groups:  map[string][]string{},
				UserIDs: map[string]string{"serial01": "10"},
			},
		},
	}

	for i, test := range tests {
		// Copy the files in configDir to a temporary directory.
		tempConfigDir, err := ioutil.TempDir("", "madbConfigTest")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempConfigDir)

		files, err := ioutil.ReadDir(test.configDir)
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range files {
			src, err := os.Open(filepath.Join(test.configDir, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			dst, err := os.Create(filepath.Join(tempConfigDir, file.Name()))
			if err != nil {
				t.Fatal(err)
			}

			_, err = io.Copy(dst, src)
			if err != nil {
				t.Fatal(err)
			}

			src.Close()
			dst.Close()
		}

		// Run the migration.
		migrateOldConfigFiles(tempConfigDir)

		// Check the resulting config
		cfg, err := readConfig(filepath.Join(tempConfigDir, "config"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := *cfg, test.want; !reflect.DeepEqual(got, want) {
			t.Fatalf("unmatched results for tests[%v]: got %v, want %v", i, got, want)
		}

		// Check the file mapping.
		for oldFile, newFile := range test.fileMap {
			if oldFile == "" {
				if _, err := os.Stat(filepath.Join(tempConfigDir, newFile)); os.IsNotExist(err) {
					t.Fatalf("missing an expected config file %q", newFile)
				}
				continue
			}

			oldBytes, err := ioutil.ReadFile(filepath.Join(test.configDir, oldFile))
			if err != nil {
				t.Fatal(err)
			}
			newBytes, err := ioutil.ReadFile(filepath.Join(tempConfigDir, newFile))
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(oldBytes, newBytes) != 0 {
				t.Fatalf("unmatched file contents between %q and %q.", oldFile, newFile)
			}
		}
	}
}
