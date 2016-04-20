// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var cmdMadbExec = &cmdline.Command{
	Runner: subCommandRunner{subCmd: runMadbExecForDevice},
	Name:   "exec",
	Short:  "Run the provided adb command on all devices and emulators concurrently",
	Long: `
Runs the provided adb command on all devices and emulators concurrently.

For example, the following line:

    madb exec push ./foo.txt /sdcard/foo.txt

copies the ./foo.txt file to /sdcard/foo.txt for all the currently connected devices.

There are a few pre-defined keywords that can be expanded within an argument.

    "{{index}}"  : the index of the current device, starting from 1.
    "{{name}}"   : the nickname of the current device, or the serial number if a nickname is not set.
    "{{serial}}" : the serial number of the current device.

For example, the following line:

    madb exec -n=Alice,Bob push ./{{name}}.txt /sdcard/{{name}}.txt

copies the ./Alice.txt file to the device named Alice, and ./Bob.txt to the device named Bob.
Note that you should type in "{{name}}" as-is, with the opening/closing curly braces, similar to
when you're using a template library such as mustache.

To see the list of available adb commands, type 'adb help'.
`,
	ArgsName: "<command>",
	ArgsLong: `
<command> is a normal adb command, which will be executed on all devices and emulators.
`,
}

func runMadbExecForDevice(env *cmdline.Env, args []string, d device, properties variantProperties) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	// Expand the keywords before running the command.
	expandedArgs := make([]string, len(args))
	for i, arg := range args {
		expandedArgs[i] = expandKeywords(arg, d)
	}

	cmdArgs := append([]string{"-s", d.Serial}, expandedArgs...)
	cmd := sh.Cmd("adb", cmdArgs...)
	return runGoshCommandForDevice(cmd, d, false)
}
