// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $JIRI_ROOT/release/go/src/v.io/x/lib/cmdline/testdata/gendoc.go .

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var (
	allDevicesFlag   bool
	allEmulatorsFlag bool
	devicesFlag      string
)

func init() {
	cmdMadb.Flags.BoolVar(&allDevicesFlag, "d", false, `Restrict the command to only run on real devices.`)
	cmdMadb.Flags.BoolVar(&allEmulatorsFlag, "e", false, `Restrict the command to only run on emulators.`)
	cmdMadb.Flags.StringVar(&devicesFlag, "n", "", `Comma-separated device serials, qualifiers, or nicknames (set by 'madb name').  Command will be run only on specified devices.`)
}

var cmdMadb = &cmdline.Command{
	Children: []*cmdline.Command{cmdMadbExec, cmdMadbStart, cmdMadbName},
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
func getDevices(nicknameFile string) ([]device, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	output := sh.Cmd("adb", "devices", "-l").Stdout()

	nsm, err := readNicknameSerialMap(nicknameFile)
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

// Gets all the devices specified by the device specifier flags.
// Intended to be used by most of the madb sub-commands except for 'madb name'.
func getSpecifiedDevices() ([]device, error) {
	allDevices, err := getDevices(getDefaultNameFilePath())
	if err != nil {
		return nil, err
	}

	filtered := filterSpecifiedDevices(allDevices)

	if len(filtered) == 0 {
		return nil, fmt.Errorf("No devices matching the device specifiers.")
	}

	return filtered, nil
}

func filterSpecifiedDevices(devices []device) []device {
	// If no device specifier flags are set, run on all devices and emulators.
	if noDevicesSpecified() {
		return devices
	}

	result := make([]device, 0, len(devices))

	for _, d := range devices {
		if shouldIncludeDevice(d) {
			result = append(result, d)
		}
	}

	return result
}

func noDevicesSpecified() bool {
	return allDevicesFlag == false &&
		allEmulatorsFlag == false &&
		devicesFlag == ""
}

func shouldIncludeDevice(d device) bool {
	if allDevicesFlag && d.Type == realDevice {
		return true
	}

	if allEmulatorsFlag && d.Type == emulator {
		return true
	}

	tokens := strings.Split(devicesFlag, ",")
	for _, token := range tokens {
		// Ignore empty tokens
		if token == "" {
			continue
		}

		if d.Serial == token || d.Nickname == token {
			return true
		}

		for _, qualifier := range d.Qualifiers {
			if qualifier == token {
				return true
			}
		}
	}

	return false
}

type subCommandFunc func(env *cmdline.Env, args []string, d device) error

type subCommandRunner struct {
	subCmd subCommandFunc
}

var _ cmdline.Runner = (*subCommandRunner)(nil)

// Invokes the sub command on all the devices in parallel.
func (r subCommandRunner) Run(env *cmdline.Env, args []string) error {
	if err := startAdbServer(); err != nil {
		return err
	}

	devices, err := getSpecifiedDevices()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	var errs []error
	var errDevices []device

	for _, d := range devices {
		// Capture the current value.
		deviceCopy := d

		wg.Add(1)
		go func() {
			// Remember the first error returned by the sub command.
			if e := r.subCmd(env, args, deviceCopy); err == nil && e != nil {
				errs = append(errs, e)
				errDevices = append(errDevices, deviceCopy)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	// Report any errors returned from the go-routines.
	if errs != nil {
		buffer := bytes.Buffer{}
		buffer.WriteString("Error occurred while running the command on the following devices:")
		for i := 0; i < len(errs); i++ {
			buffer.WriteString("\n[" + errDevices[i].displayName() + "]\t" + errs[i].Error())
		}
		return fmt.Errorf(buffer.String())
	}

	return nil
}
