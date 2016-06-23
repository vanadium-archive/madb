// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"v.io/x/lib/cmdline"
)

var cmdMadbResolve = &cmdline.Command{
	Runner:           subCommandRunnerWithFilepath{runMadbResolve, getDefaultConfigFilePath},
	Name:             "resolve",
	DontInheritFlags: true,
	Short:            "Resolve device specifiers into device serials",
	Long: `
Resolves the provided device specifiers and prints out their device serials,
each in a separate line. This command only displays the unique serials of the
devices that are currently available.

This command can be useful when you want to use the device nicknames and groups
defined by madb in other command line tools. For example, to run a flutter app
on "MyTablet" device, you can use the following command (in Bash):

    flutter run --device $(madb resolve MyTablet)
`,
	ArgsName: "<specifier1> [<specifier2> ...]",
	ArgsLong: `
<specifier> can be anything that is accepted in the '-n' flag (see 'madb help').
It can be a device serial, qualifier, index, nickname, or a device group name.
`,
}

func runMadbResolve(env *cmdline.Env, args []string, filename string) error {
	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}

	devices, err := getDevices(cfg)
	if err != nil {
		return err
	}

	filtered, err := filterSpecifiedDevices(devices, cfg, false, false, args)
	if err != nil {
		return err
	}

	for _, d := range filtered {
		fmt.Println(d.Serial)
	}

	return nil
}
