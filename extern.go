// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var cmdMadbExtern = &cmdline.Command{
	Runner: subCommandRunner{subCmd: runMadbExternForDevice},
	Name:   "extern",
	Short:  "Run the provided external command for all devices",
	Long: `
Runs the provided external command for all devices and emulators concurrently.

For each available device, this command will spawn a sub-shell with the
ANDROID_SERIAL environmental variable set to the target device serial, and then
will run the provided external command.

There are a few pre-defined keywords that can be expanded within an argument.

    "{{index}}"  : the index of the current device, starting from 1.
    "{{name}}"   : the nickname of the current device, or the serial number if a nickname is not set.
    "{{serial}}" : the serial number of the current device.

For example, the following line:

    madb extern echo I am {{name}}, and my serial number is {{serial}}.

prints out the name and serial pairs for each device.

Note that you should type in "{{name}}" as-is, with the opening/closing curly
braces, similar to when you're using a template library such as mustache.

This command is intended to be used with external commands that are designed to
work with only a single device at a time (e.g. gomobile, flutter).
`,
	ArgsName: "<external_command>",
	ArgsLong: `
<external_command> is an external shell command to run for all devices and emulators.
`,
}

func runMadbExternForDevice(env *cmdline.Env, args []string, d device, properties variantProperties) error {
	return runExternalCommandForDevice(env, args, d, properties)
}

func runExternalCommandForDevice(env *cmdline.Env, args []string, d device, properties variantProperties) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	// Expand the keywords before running the command.
	cmdArgs := make([]string, len(args))
	for i, arg := range args {
		cmdArgs[i] = expandKeywords(arg, d)
	}

	// Set the ANDROID_SERIAL variable.
	sh.Vars["ANDROID_SERIAL"] = d.Serial

	cmd := sh.Cmd(cmdArgs[0], cmdArgs[1:]...)
	return runGoshCommandForDevice(cmd, d, false)
}
