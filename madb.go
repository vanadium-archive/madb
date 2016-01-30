// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $JIRI_ROOT/release/go/src/v.io/x/lib/cmdline/testdata/gendoc.go .

package main

import (
	"fmt"
	"os"
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

type deviceType string

const (
	emulator   deviceType = "Emulator"
	realDevice deviceType = "RealDevice"
)

type device struct {
	Serial     string
	Type       deviceType
	Qualifiers []string
	Nickname   string
}

// Returns the display name which is intended to be used as the console output prefix.
// This would be the nickname of the device if there is one; otherwise, the serial number is used.
func (d device) displayName() string {
	if d.Nickname != "" {
		return d.Nickname
	}

	return d.Serial
}

// Runs "adb devices -l" command, and parses the result to get all the device serial numbers.
func getDevices(filename string) ([]device, error) {
	sh := gosh.NewShell(gosh.Opts{})
	defer sh.Cleanup()

	output := sh.Cmd("adb", "devices", "-l").Stdout()

	nsm, err := readNicknameSerialMap(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not read the nickname file.")
	}

	return parseDevicesOutput(output, nsm)
}

// Parses the output generated from "adb devices -l" command and return the list of device serial numbers
// Devices that are currently offline are excluded from the returned list.
func parseDevicesOutput(output string, nsm map[string]string) ([]device, error) {
	lines := strings.Split(output, "\n")

	result := []device{}

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

		// Fill in the device serial and all the qualifiers.
		d := device{
			Serial:     fields[0],
			Qualifiers: fields[2:],
		}

		// Determine whether this device is an emulator or a real device.
		if strings.HasPrefix(d.Serial, "emulator") {
			d.Type = emulator
		} else {
			d.Type = realDevice
		}

		// Determine whether there is a nickname defined for this device,
		// so that the console output prefix can display the nickname instead of the serial.
	NSMLoop:
		for nickname, serial := range nsm {
			if d.Serial == serial {
				d.Nickname = nickname
				break
			}

			for _, qualifier := range d.Qualifiers {
				if qualifier == serial {
					d.Nickname = nickname
					break NSMLoop
				}
			}
		}

		result = append(result, d)
	}

	return result, nil
}
