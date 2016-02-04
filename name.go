// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"v.io/x/lib/cmdline"
)

var cmdMadbName = &cmdline.Command{
	Children: []*cmdline.Command{cmdMadbNameSet, cmdMadbNameUnset, cmdMadbNameList, cmdMadbNameClearAll},
	Name:     "name",
	Short:    "Manage device nicknames",
	Long: `
Manages device nicknames, which are meant to be more human-friendly compared to
the device serials provided by adb tool.

NOTE: Device specifier flags (-d, -e, -n) are ignored in all 'madb name' commands.
`,
}

var cmdMadbNameSet = &cmdline.Command{
	Runner: runnerFuncWithFilepath(runMadbNameSet),
	Name:   "set",
	Short:  "Set a nickname to be used in place of the device serial.",
	Long: `
Sets a human-friendly nickname that can be used when specifying the device in
any madb commands.

The device serial can be obtain using the 'adb devices -l' command.
For example, consider the following example output:

    HT4BVWV00023           device usb:3-3.4.2 product:volantisg model:Nexus_9 device:flounder_lte

The first value, 'HT4BVWV00023', is the device serial.
To assign a nickname for this device, run the following command:

    madb name set HT4BVWV00023 MyTablet

and it will assign the 'MyTablet' nickname to the device serial 'HT4BVWV00023'.
The alternative device specifiers (e.g., 'usb:3-3.4.2', 'product:volantisg')
can also have nicknames.

When a nickname is set for a device serial, the nickname can be used to specify
the device within madb commands.

There can only be one nickname for a device serial.
When the 'madb name set' command is invoked with a device serial with an already
assigned nickname, the old one will be replaced with the newly provided one.
`,
	ArgsName: "<device_serial> <nickname>",
	ArgsLong: `
<device_serial> is a device serial (e.g., 'HT4BVWV00023') or an alternative device specifier (e.g., 'usb:3-3.4.2') obtained from 'adb devices -l' command
<nickname> is an alpha-numeric string with no special characters or spaces.
`,
}

func runMadbNameSet(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) != 2 {
		return env.UsageErrorf("There must be exactly two arguments.")
	}

	serial, nickname := args[0], args[1]
	if !isValidDeviceSerial(serial) {
		return env.UsageErrorf("Not a valid device serial: %v", serial)
	}

	if !isValidNickname(nickname) {
		return env.UsageErrorf("Not a valid nickname: %v", nickname)
	}

	nsm, err := readNicknameSerialMap(filename)
	if err != nil {
		return err
	}

	// If the nickname is already in use, don't allow it at all.
	if _, present := nsm[nickname]; present {
		return fmt.Errorf("The provided nickname %q is already in use.", nickname)
	}

	// If the serial number already has an assigned nickname, delete it first.
	// Need to do this check, because the nickname-serial map should be a one-to-one mapping.
	if nn, present := reverseMap(nsm)[serial]; present {
		delete(nsm, nn)
	}

	// Add the nickname serial mapping.
	nsm[nickname] = serial

	return writeNicknameSerialMap(nsm, filename)
}

var cmdMadbNameUnset = &cmdline.Command{
	Runner: runnerFuncWithFilepath(runMadbNameUnset),
	Name:   "unset",
	Short:  "Unset a nickname set by the 'madb name set' command.",
	Long: `
Unsets a nickname assigned by the 'madb name set' command. Either the device
serial or the assigned nickname can be specified to remove the mapping.
`,
	ArgsName: "<device_serial | nickname>",
	ArgsLong: `
There should be only one argument, which is either the device serial or the nickname.
`,
}

func runMadbNameUnset(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) != 1 {
		return env.UsageErrorf("There must be exactly one argument.")
	}

	name := args[0]
	if !isValidDeviceSerial(name) && !isValidNickname(name) {
		return env.UsageErrorf("Not a valid device serial or name: %v", name)
	}

	nsm, err := readNicknameSerialMap(filename)
	if err != nil {
		return err
	}

	found := false
	for nickname, serial := range nsm {
		if nickname == name || serial == name {
			delete(nsm, nickname)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("The provided argument is neither a known nickname nor a device serial.")
	}

	return writeNicknameSerialMap(nsm, filename)
}

var cmdMadbNameList = &cmdline.Command{
	Runner: runnerFuncWithFilepath(runMadbNameList),
	Name:   "list",
	Short:  "List all the existing nicknames.",
	Long: `
Lists all the currently stored nicknames of device serials.
`,
}

func runMadbNameList(env *cmdline.Env, args []string, filename string) error {
	nsm, err := readNicknameSerialMap(filename)
	if err != nil {
		return err
	}

	// TODO(youngseokyoon): pretty print this.
	fmt.Println("Serial          Nickname")
	fmt.Println("========================")

	for nickname, serial := range nsm {
		fmt.Printf("%v\t%v\n", serial, nickname)
	}

	return nil
}

var cmdMadbNameClearAll = &cmdline.Command{
	Runner: runnerFuncWithFilepath(runMadbNameClearAll),
	Name:   "clear-all",
	Short:  "Clear all the existing nicknames.",
	Long: `
Clears all the currently stored nicknames of device serials.
`,
}

func runMadbNameClearAll(env *cmdline.Env, args []string, filename string) error {
	return os.Remove(filename)
}

func getDefaultNameFilePath() string {
	return filepath.Join(os.Getenv("HOME"), ".madb_names")
}

func isValidDeviceSerial(serial string) bool {
	r := regexp.MustCompile(`^([A-Za-z0-9:\-\._]+|@\d+)$`)
	return r.MatchString(serial)
}

func isValidNickname(nickname string) bool {
	r := regexp.MustCompile(`^\w+$`)
	return r.MatchString(nickname)
}

// reverseMap returns a new map which contains reversed key, value pairs in the original map.
// The source map is assumed to be a one-to-one mapping between keys and values.
func reverseMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}

	reversed := make(map[string]string, len(source))
	for k, v := range source {
		reversed[v] = k
	}

	return reversed
}

// readNicknameSerialMap reads the provided file and reconstructs the nickname => serial map.
// The mapping is written one per each line, in the form of "<nickname> <serial>".
func readNicknameSerialMap(filename string) (map[string]string, error) {
	result := make(map[string]string)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// TODO(youngseokyoon): use encoding/json for serialization and deserialization.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 2 {
			result[fields[0]] = fields[1]
		} else {
			return nil, fmt.Errorf("Unexpected number of columns in the nickname file.")
		}
	}

	return result, nil
}

// writeNicknameSerialmap takes a nickname => serial map and writes it into the provided file name.
// The mapping is written one per each line, in the form of "<nickname> <serial>".
func writeNicknameSerialMap(nsm map[string]string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for nickname, serial := range nsm {
		fmt.Fprintln(f, nickname, serial)
	}

	return nil
}

// runnerFuncWithFilepath is an adapter that turns the madb name subcommand functions into cmdline.Runners.
type runnerFuncWithFilepath func(*cmdline.Env, []string, string) error

// Run implements the cmdline.Runner interface by providing the default name file path
// as the third string argument of the underlying run function.
func (f runnerFuncWithFilepath) Run(env *cmdline.Env, args []string) error {
	return f(env, args, getDefaultNameFilePath())
}
