// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var (
	forceStopFlag bool
)

func init() {
	initializeIDCacheFlags(&cmdMadbStart.Flags)
	cmdMadbStart.Flags.BoolVar(&forceStopFlag, "force-stop", true, `Force stop the target app before starting the activity.`)
}

var cmdMadbStart = &cmdline.Command{
	Runner: subCommandRunner{initMadbStart, runMadbStartForDevice},
	Name:   "start",
	Short:  "Launch your app on all devices",
	Long: `
Launches your app on all devices.

`,
	ArgsName: "[<application_id> <activity_name>]",
	ArgsLong: `
<application_id> is usually the package name where the activities are defined.
(See: http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

<activity_name> is the Java class name for the activity you want to launch.
If the package name of the activity is different from the application ID, the activity name must be
a fully-qualified name (e.g., com.yourcompany.yourapp.MainActivity).

If either <application_id> or <activity_name> is provided, the other must be provided as well.


If no arguments are specified, madb automatically determines which app to launch, based on the build
scripts found in the current working directory.

1) If the working directory contains a Flutter project (i.e., has "flutter.yaml"), this command will
run "flutter start --android-device-id=<device serial>" for all the specified devices.

2) If the working directory contains a Gradle Android project (i.e., has "build.gradle"), this
command will run a small Gradle script to extract the application ID and the main activity name.
In this case, the extracted IDs are cached, so that "madb start" can be repeated without even
running the Gradle script again.  The IDs can be re-extracted by clearing the cache by providing
"-clear-cache" flag.
`,
}

func initMadbStart(env *cmdline.Env, args []string) ([]string, error) {
	return initMadbCommand(env, args, true, true)
}

func runMadbStartForDevice(env *cmdline.Env, args []string, d device) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	if len(args) == 2 {
		appID, activity := args[0], args[1]

		// In case the activity name is a simple name (i.e. without the package name), add a dot in the front.
		// This is a shorthand syntax to prepend the activity name with the package name provided in the manifest.
		// http://developer.android.com/guide/topics/manifest/activity-element.html#nm
		if !strings.ContainsAny(activity, ".") {
			activity = "." + activity
		}

		// More details on the "adb shell am" command can be found at: http://developer.android.com/tools/help/shell.html#am
		cmdArgs := []string{"-s", d.Serial, "shell", "am", "start"}
		if forceStopFlag {
			cmdArgs = append(cmdArgs, "-S")
		}
		cmdArgs = append(cmdArgs, "-n", appID+"/"+activity)
		cmd := sh.Cmd("adb", cmdArgs...)
		return runGoshCommandForDevice(cmd, d)
	}

	// In case of flutter, the application ID is not even needed.
	// Simply run "flutter start --android-device-id <device_serial>" on all devices.
	if isFlutterProject(wd) {
		cmdArgs := []string{"start", "--android-device-id", d.Serial}
		cmd := sh.Cmd("flutter", cmdArgs...)
		return runGoshCommandForDevice(cmd, d)
	}

	return fmt.Errorf("No arguments are provided and failed to extract the ids from the build scripts.")
}
