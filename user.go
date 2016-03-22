// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"v.io/x/lib/cmdline"
)

// TODO(youngseokyoon): add a helper command that wraps "madb exec shell pm list users" to show all available users.
var cmdMadbUser = &cmdline.Command{
	Children: []*cmdline.Command{cmdMadbUserSet, cmdMadbUserUnset, cmdMadbUserList, cmdMadbUserClearAll},
	Name:     "user",
	Short:    "Manage default user settings for each device",
	Long: `
Manages default user settings for each device.

An Android device can have multiple user accounts, and each user account has a numeric ID associated
with it. Certain adb commands accept '--user <user_id>' as a parameter to allow specifying which of
the Android user account should be used when running the command. The default behavior when the
user ID is not provided varies by the adb command being run.

Some madb commands internally run these adb commands which accept the '--user' flag. You can let
madb use different user IDs for different devices by storing the default user ID for each device
using 'madb user set' command. If the default user ID is not set for a particular device, madb will
not provide the '--user' flag to the underlying adb command, and the current user will be used for
that device as a result.

Below is the list of madb commands which are affected by the default user ID settings:

    madb clear-data
    madb start
    madb stop
    madb uninstall

For more details on how to obtain the user ID from an Android device, see 'madb user help set'.

NOTE: Device specifier flags (-d, -e, -n) are ignored in all 'madb name' commands.
`,
}

var cmdMadbUserSet = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbUserSet, getDefaultUserFilePath},
	Name:   "set",
	Short:  "Set a default user ID to be used for the given device.",
	Long: `
Sets a default user ID to be used for the specified device, when there are multiple user accounts on
a single device.

The user IDs can be obtained using the 'adb [<device_serial>] shell pm list users' command.
Alternatively, you can use 'madb exec' if you want to specify the device with a nickname.
For example, running the following command:

    madb -n=MyPhone exec shell pm list users

will list the available users and their IDs on the MyPhone device.
Consider the following example output:

    [MyPhone]       Users:
    [MyPhone]               UserInfo{0:John Doe:13} running
    [MyPhone]               UserInfo{10:Work profile:30} running

There are two available users, "John Doe" and "Work profile". Each user is assigned a "user ID",
which appears on the left of the user name. In this case, the user ID of "John Doe" is "0", and the
user ID of the "Work profile" is "10".

To use the "Work profile" as the default user when running madb commands on this device, run the
following command:

    madb user set MyPhone 10

and then madb will use "Work profile" as the default user for device "MyPhone" in any of the
subsequence madb commands.
`,
	ArgsName: "<device_serial> <user_id>",
	ArgsLong: `
<device_serial> is the unique serial number for the device, which can be obtained from 'adb devices'.
<user_id> is one of the user IDs obtained from 'adb shell pm list users' command.
`,
}

func runMadbUserSet(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) != 2 {
		return fmt.Errorf("There must be exactly two arguments.")
	}

	// TODO(youngseokyoon): make it possible to specify the device using its nickname or index.
	// Validate the device serial
	serial := args[0]
	if !isValidDeviceSerial(serial) {
		return fmt.Errorf("Not a valid device serial: %v", serial)
	}

	// Validate the user ID.
	userID := args[1]
	if id, err := strconv.Atoi(userID); err != nil || id < 0 {
		return fmt.Errorf("Not a valid user ID: %v", userID)
	}

	// Get the <device_serial, user_id> mapping.
	serialUserMap, err := readMapFromFile(filename)
	if err != nil {
		return err
	}

	// Add the <device_serial, user_id> mapping for the specified device.
	serialUserMap[serial] = userID
	return writeMapToFile(serialUserMap, filename)
}

var cmdMadbUserUnset = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbUserUnset, getDefaultUserFilePath},
	Name:   "unset",
	Short:  "Unset the default user ID set by the 'madb user set' command.",
	Long: `
Unsets the default user ID assigned by the 'madb user set' command for the specified device.

Running this command without any device specifiers will unset the default users only for the
currently available devices and emulators, while keeping the default user IDs for the other devices.
`,
	ArgsName: "<device_serial>",
	ArgsLong: `
<device_serial> is the unique serial number for the device, which can be obtained from 'adb devices'.
`,
}

func runMadbUserUnset(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) != 1 {
		return fmt.Errorf("There must be exactly one argument.")
	}

	// TODO(youngseokyoon): make it possible to specify the device using its nickname or index.
	// Validate the device serial
	serial := args[0]
	if !isValidDeviceSerial(serial) {
		return fmt.Errorf("Not a valid device serial: %v", serial)
	}

	// Get the <device_serial, user_id> mapping.
	serialUserMap, err := readMapFromFile(filename)
	if err != nil {
		return err
	}

	// Delete the <device_serial, user_id> mapping for the specified device.
	delete(serialUserMap, serial)
	return writeMapToFile(serialUserMap, filename)
}

var cmdMadbUserList = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbUserList, getDefaultUserFilePath},
	Name:   "list",
	Short:  "List all the existing default user IDs.",
	Long: `
Lists all the currently stored default user IDs for devices.
`,
}

func runMadbUserList(env *cmdline.Env, args []string, filename string) error {
	// Get the <device_serial, user_id> mapping.
	serialUserMap, err := readMapFromFile(filename)
	if err != nil {
		return err
	}

	// TODO(youngseokyoon): pretty print this.
	fmt.Println("Device Serial    User ID")
	fmt.Println("========================")

	for s, u := range serialUserMap {
		fmt.Printf("%v\t%v\n", s, u)
	}

	return nil
}

var cmdMadbUserClearAll = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbUserClearAll, getDefaultUserFilePath},
	Name:   "clear-all",
	Short:  "Clear all the existing default user settings.",
	Long: `
Clears all the currently stored default user IDs for devices.

This command clears the default user IDs regardless of whether the device is currently connected or not.
`,
}

func runMadbUserClearAll(env *cmdline.Env, args []string, filename string) error {
	return os.Remove(filename)
}

func getDefaultUserFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "users"), nil
}
