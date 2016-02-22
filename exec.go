// Copyright 2015 The Vanadium Authors. All rights reserved.
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

    madb -a exec push ./foo.txt /sdcard/foo.txt

copies the ./foo.txt file to /sdcard/foo.txt for all the currently connected Android devices.

To see the list of available adb commands, type 'adb help'.
`,
	ArgsName: "<command>",
	ArgsLong: `
<command> is a normal adb command, which will be executed on all devices and emulators.
`,
}

func runMadbExecForDevice(env *cmdline.Env, args []string, d device) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	cmdArgs := append([]string{"-s", d.Serial}, args...)
	cmd := sh.Cmd("adb", cmdArgs...)
	return runGoshCommandForDevice(cmd, d)
}
