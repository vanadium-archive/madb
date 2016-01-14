// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $JIRI_ROOT/release/go/src/v.io/x/lib/cmdline/testdata/gendoc.go .

package main

import (
	"fmt"
	"os/exec"
	"strings"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var cmdMadb = &cmdline.Command{
	Children: []*cmdline.Command{cmdMadbExec, cmdMadbName},
	Name:     "madb",
	Short:    "Multi-device Android Debug Bridge",
	Long: `
Multi-device Android Debug Bridge

The madb command wraps Android Debug Bridge (adb) command line tool
and provides various features for controlling multiple Android devices concurrently.
`,
}

func main() {
	cmdline.Main(cmdMadb)
}

// Makes sure that adb server is running.
// Intended to be called at the beginning of each subcommand.
func startAdbServer() error {
	// TODO(youngseokyoon): search for installed adb tool more rigourously.
	if err := exec.Command("adb", "start-server").Run(); err != nil {
		return fmt.Errorf("Failed to start adb server. Please make sure that adb is in your PATH: %v", err)
	}

	return nil
}

// Runs "adb devices" command, and parses the result to get all the device serial numbers.
func getDevices() ([]string, error) {
	sh := gosh.NewShell(gosh.Opts{})
	defer sh.Cleanup()

	output := sh.Cmd("adb", "devices", "-l").Stdout()

	return parseDevicesOutput(output)
}

// Parses the output generated from "adb devices -l" command and return the list of device serial numbers
// Devices that are currently offline are excluded from the returned list.
func parseDevicesOutput(output string) ([]string, error) {
	lines := strings.Split(output, "\n")

	result := []string{}

	// Check the first line of the output
	if len(lines) <= 0 || strings.TrimSpace(lines[0]) != "List of devices attached" {
		return result, fmt.Errorf("The output from 'adb devices -l' command does not look as expected.")
	}

	// Iterate over all the device serial numbers, starting from the second line.
	for _, line := range lines[1:] {
		fields := strings.Fields(line)

		if len(fields) <= 1 || fields[1] == "offline" {
			continue
		}

		result = append(result, strings.Fields(line)[0])
	}

	return result, nil
}
