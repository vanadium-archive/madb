// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// A script for creating pre-compiled binary releases for multiple platforms.
// It reads the MADB_VERSION number,
//
// This script is intended to be run from the parent directory with the following command:
//
//     go run scripts/release.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"v.io/x/lib/gosh"
)

var targetOS = map[string]string{
	"linux":  "linux",
	"darwin": "macosx",
}

var targetArch = []string{"amd64"}

func main() {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	// Read the version string from the file MADB_VERSION.
	versionString, err := readVersionString()
	if err != nil {
		panic(err)
	}

	archiveDir := "archive"
	if err := os.Mkdir(archiveDir, 0755); err != nil {
		panic(err)
	}

	for osCode, osName := range targetOS {
		for _, arch := range targetArch {
			variantName := fmt.Sprintf("madb-%v-%v-%v", versionString, osName, arch)
			tempDir := sh.MakeTempDir()
			outputDir := filepath.Join(tempDir, variantName)
			outputPath := filepath.Join(outputDir, "madb")

			// The -ldflags "-X main.version=<versionString>" flag overwrites the version value
			// declared in version.go.
			cmd := sh.Cmd("jiri", "go", "build", "-o", outputPath, "-ldflags", fmt.Sprintf("-X main.version=%v", versionString))
			cmd.Vars["GOOS"] = osCode
			cmd.Vars["GOARCH"] = arch
			cmd.Run()

			// Archive the directory using the pax utility.
			archivePath := filepath.Join(archiveDir, variantName+".tar.gz")
			cmd = sh.Cmd("pax", "-w", "-z", "-M", "dist", "-s", "#^"+tempDir+"/##", "-f", archivePath, outputDir)
			cmd.Run()
		}
	}
}

func readVersionString() (string, error) {
	versionFile, err := os.Open("MADB_VERSION")
	if err != nil {
		return "", err
	}
	defer versionFile.Close()

	bytes, err := ioutil.ReadAll(versionFile)
	return strings.TrimSpace(string(bytes)), err
}
