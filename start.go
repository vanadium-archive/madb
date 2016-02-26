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
	"v.io/x/lib/textutil"
)

var cmdMadbStart = &cmdline.Command{
	Runner: subCommandRunner{runMadbStartForDevice},
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

func runMadbStartForDevice(env *cmdline.Env, args []string, d device) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	if len(args) != 2 {
		// TODO(youngseokyoon): Extract the application ID and activity name from the build scripts in the current directory.
		return fmt.Errorf("You must provide the application ID and the activity name.")
	}

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

	stdout := textutil.PrefixLineWriter(os.Stdout, "["+d.displayName()+"]\t")
	stderr := textutil.PrefixLineWriter(os.Stderr, "["+d.displayName()+"]\t")
	cmd.AddStdoutWriter(stdout)
	cmd.AddStderrWriter(stderr)
	cmd.Run()
	stdout.Flush()
	stderr.Flush()

	return nil
}
