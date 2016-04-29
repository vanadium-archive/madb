// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"v.io/x/lib/cmdline"
)

var cmdMadbVersion = &cmdline.Command{
	Runner:           cmdline.RunnerFunc(runCmdMadbVersion),
	Name:             "version",
	DontInheritFlags: true,
	Short:            "Print the madb version number",
	Long: `
Prints the madb version number to the console.

If this version of madb binary is an official release, this command will show the version number.
Otherwise, the version will be in the form of "<version>-develop", where the version indicates the
most recent stable release version prior to this version of madb binary.
`,
}

func runCmdMadbVersion(env *cmdline.Env, args []string) error {
	fmt.Println("madb version:", version)
	return nil
}
