// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
	"v.io/x/lib/textutil"
)

var cmdMadbExec = &cmdline.Command{
	Runner: subCommandRunner{runMadbExecForDevice},
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
	sh := gosh.NewShell(gosh.Opts{})
	defer sh.Cleanup()

	cmdArgs := append([]string{"-s", d.Serial}, args...)
	cmd := sh.Cmd("adb", cmdArgs...)

	stdout := textutil.PrefixLineWriter(os.Stdout, "["+d.displayName()+"]\t")
	stderr := textutil.PrefixLineWriter(os.Stderr, "["+d.displayName()+"]\t")
	cmd.AddStdoutWriter(stdout)
	cmd.AddStderrWriter(stderr)
	cmd.Run()
	stdout.Flush()
	stderr.Flush()

	return sh.Err
}
