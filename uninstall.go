// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var (
	keepDataFlag bool
)

func init() {
	initializeIDCacheFlags(&cmdMadbUninstall.Flags)
	cmdMadbUninstall.Flags.BoolVar(&keepDataFlag, "keep-data", false, `Keep the application data and cache directories.  Equivalent to '-k' flag in 'adb uninstall' command.`)
}

var cmdMadbUninstall = &cmdline.Command{
	Runner: subCommandRunner{initMadbUninstall, runMadbUninstallForDevice},
	Name:   "uninstall",
	Short:  "Uninstall your app from all devices",
	Long: `
Uninstall your app from all devices.

`,
	ArgsName: "[<application_id>]",
	ArgsLong: `
<application_id> is usually the package name where the activities are defined.
(See: http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)


If the application_id is not specified, madb automatically determines which app to uninstall, based
on the build scripts found in the current working directory.

If the working directory contains a Gradle Android project (i.e., has "build.gradle"), run a small
Gradle script to extract the application ID.  In this case, the extracted ID is cached, so that
"madb uninstall" can be repeated without even running the Gradle script again.  The ID can be
re-extracted by clearing the cache by providing "-clear-cache" flag.
`,
}

func initMadbUninstall(env *cmdline.Env, args []string) ([]string, error) {
	return initMadbCommand(env, args, false, false)
}

func runMadbUninstallForDevice(env *cmdline.Env, args []string, d device) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	if len(args) == 1 {
		appID := args[0]

		cmdArgs := []string{"-s", d.Serial, "uninstall"}
		if keepDataFlag {
			cmdArgs = append(cmdArgs, "-k")
		}
		cmdArgs = append(cmdArgs, appID)
		cmd := sh.Cmd("adb", cmdArgs...)
		return runGoshCommandForDevice(cmd, d)
	}

	return fmt.Errorf("No arguments are provided and failed to extract the id from the build scripts.")
}
