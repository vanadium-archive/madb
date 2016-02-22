// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var (
	wd string // working directory
)

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

var cmdMadbStart = &cmdline.Command{
	Runner: subCommandRunner{initMadbStart, runMadbStartForDevice},
	Name:   "start",
	Short:  "Launch your app on all devices",
	Long: `
Launches your app on all devices.

`,
	ArgsName: "<application_id> <activity_name>",
	ArgsLong: `
<application_id> is usually the package name where the activities are defined.
(See: http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

<activity_name> is the Java class name for the activity you want to launch.
If the package name of the activity is different from the application ID, the activity name must be a fully-qualified name (e.g., com.yourcompany.yourapp.MainActivity).
`,
}

func initMadbStart(env *cmdline.Env, args []string) ([]string, error) {
	// If both arguments are provided, or if it is a flutter project, simply pass the arguments through.
	if len(args) == 2 || isFlutterProject(wd) {
		return args, nil
	}

	if len(args) != 0 {
		return nil, fmt.Errorf("You mush provide either zero or exactly two arguments.")
	}

	// Try to extract the application ID and the main activity name from the Gradle scripts.
	// TODO(youngseokyoon): cache the ids, since the ids are not supposed to be changed very often.
	if isGradleProject(wd) {
		fmt.Println("Running Gradle to extract the application ID and the main activity name...")

		appID, activity, err := extractIdsFromGradle(wd)
		if err != nil {
			return nil, err
		}

		args = []string{appID, activity}
	}

	return args, nil
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

		// TODO(youngseokyoon): add a flag for not stopping the activity when it is currently running.
		// More details on the "adb shell am" command can be found at: http://developer.android.com/tools/help/shell.html#am
		cmdArgs := []string{"-s", d.Serial, "shell", "am", "start", "-S", "-n", appID + "/" + activity}
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
