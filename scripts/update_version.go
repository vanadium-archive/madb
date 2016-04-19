// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// A script used to update the version number in version.go, based on MADB_VERSION.
// For example, when the MADB_VERSION file contains "v1.0.0", the value of the version variable
// defined in version.go file should be updated to "v1.0.0-develop".
//
// This script is intended to be run as part of "go generate" from the parent directory.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const (
	tmpl = `// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated via go generate.
// To see how this file is generated, please refer to "scripts/update_version.go" file.
// DO NOT UPDATE MANUALLY

package main

// version contains the current version number of madb tool to be shown to the user.
// By default, the string "develop" will be used to tell that the current binary is a development
// version, for example when the user gets this tool with 'go get' command.
// This value is intended to be overwritten by the command line argument of 'go build', when
// releasing an official version. See "scripts/release.go" for more details.
var version = "{{.VersionString}}-develop"
`
)

const (
	versionFile      = "MADB_VERSION"
	versionGoSrcFile = "version.go"
)

func main() {
	if err := updateVersion(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update the version number: %v", err)
		os.Exit(1)
	}
}

func updateVersion() error {
	// Read the version string from the file MADB_VERSION.
	versionString, err := readFile(versionFile)
	if err != nil {
		return err
	}
	versionString = strings.TrimSpace(versionString)

	// Load the template.
	t, err := template.New("version").Parse(tmpl)
	if err != nil {
		return err
	}

	// Create the destination file.
	destFile, err := os.Create(versionGoSrcFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Define the data to be used within the template.
	data := map[string]string{"VersionString": versionString}

	// Execute the template with the above data.
	if err := t.Execute(destFile, data); err != nil {
		return err
	}

	return nil
}

func readFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	return string(bytes), err
}
