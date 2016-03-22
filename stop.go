// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

func init() {
	initializePropertyCacheFlags(&cmdMadbStop.Flags)
}

var cmdMadbStop = &cmdline.Command{
	Runner: subCommandRunner{initMadbStart, runMadbStartForDevice},
	Name:   "stop",
	Short:  "Stop your app on all devices",
	Long: `
Stops your app on all devices.

To stop your app for a specific user on a particular device, use 'madb user set' command to set the
default user ID for that device. (See 'madb help user' for more details.)

`,
	ArgsName: "[<application_id>]",
	ArgsLong: `
<application_id> is usually the package name where the activities are defined.
(See: http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)


If the application ID is not specified, madb automatically determines which app to stop, based on
the build scripts found in the current working directory.

1) If the working directory contains a Flutter project (i.e., has "flutter.yaml"), this command will
run "flutter stop --android-device-id=<device serial>" for all the specified devices.

2) If the working directory contains a Gradle Android project (i.e., has "build.gradle"), run a
small Gradle script to extract the application ID. In this case, the extracted ID is cached, so
that "madb stop" can be repeated without even running the Gradle script again. The ID can be
re-extracted by clearing the cache by providing "-clear-cache" flag.
`,
}

func initMadbStop(env *cmdline.Env, args []string) ([]string, error) {
	return initMadbCommand(env, args, true, false)
}

func runMadbStopForDevice(env *cmdline.Env, args []string, d device) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	if len(args) == 2 {
		appID := args[0]

		// More details on the "adb shell am" command can be found at: http://developer.android.com/tools/help/shell.html#am
		cmdArgs := []string{"-s", d.Serial, "shell", "am", "force-stop"}

		// Specify the user ID if applicable.
		if d.UserID != "" {
			cmdArgs = append(cmdArgs, "--user", d.UserID)
		}

		cmdArgs = append(cmdArgs, appID)
		cmd := sh.Cmd("adb", cmdArgs...)
		return runGoshCommandForDevice(cmd, d, true)
	}

	// In case of flutter, the application ID is not even needed.
	// Simply run "flutter stop --android-device-id <device_serial>" on all devices.
	if isFlutterProject(wd) {
		cmdArgs := []string{"stop", "--android-device-id", d.Serial}
		cmd := sh.Cmd("flutter", cmdArgs...)
		return runGoshCommandForDevice(cmd, d, false)
	}

	return fmt.Errorf("No arguments are provided and failed to extract the id from the build scripts.")
}
