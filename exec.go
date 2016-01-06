// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var cmdMadbExec = &cmdline.Command{
	Runner: cmdline.RunnerFunc(runMadbExec),
	Name:   "exec",
	Short:  "Run the provided adb command on all the specified devices concurrently",
	Long: `
Runs the provided adb command on all the specified devices concurrently.

For example, the following line:

    madb -a exec push ./foo.txt /sdcard/foo.txt

copies the ./foo.txt file to /sdcard/foo.txt for all the currently connected Android devices (specified by -a flag).

To see the list of available adb commands, type 'adb help'.
`,
	ArgsName: "<command>",
	ArgsLong: `
<command> is a normal adb command, which will be executed on all the specified devices.
`,
}

func runMadbExec(env *cmdline.Env, args []string) error {
	if err := startAdbServer(); err != nil {
		return err
	}

	devices, err := getDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		sh := gosh.NewShell(gosh.Opts{})
		defer sh.Cleanup()

		cmdArgs := append([]string{"-s", device}, args...)
		cmd := sh.Cmd("adb", cmdArgs...)

		// TODO(youngseokyoon): use pipes instead to prefix console messages with their device names.
		// For now, just forward all the messages to stdout/stderr.
		cmd.AddStdoutWriter(gosh.NopWriteCloser(os.Stdout))
		cmd.AddStderrWriter(gosh.NopWriteCloser(os.Stderr))

		cmd.Start()
		defer cmd.Wait()

		// TODO(youngseokyoon): check for exit code of each command
	}

	return nil
}
