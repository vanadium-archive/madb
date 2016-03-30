// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A script that takes a source Gradle script, and writes a .go file that contains a constant string
// variable that holds the contents of the source script. By doing this, the Gradle script can be
// embedded in the madb binary. This script is located under testdata, to avoid being installed in
// the $GOPATH/bin directory.
//
// This script is meant to be run via go generate from the parent directory.
// See the go:generate comment at the top of madb.go file.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

const (
	tmpl = `// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated via go generate.
// DO NOT UPDATE MANUALLY

package main

const {{.VarName}} = {{.Backtick}}{{.Contents}}{{.Backtick}}
`
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run embed_gradle_script.go <source file path> <destination .go file path> <variable name>")
		os.Exit(1)
	}

	if err := generateEmbeddedScript(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate the embedded script: %v", err)
		os.Exit(1)
	}
}

func generateEmbeddedScript(source, dest, varName string) error {
	// Read the source file.
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	bytes, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return err
	}

	// Create the destination file.
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Load the template.
	t, err := template.New("embedded_gradle").Parse(tmpl)
	if err != nil {
		return err
	}

	// Define the data to be used within the template.
	data := map[string]string{
		"VarName":  varName,
		"Contents": string(bytes),
		"Backtick": "`",
	}

	// Execute the template with the above data.
	if err := t.Execute(destFile, data); err != nil {
		return err
	}

	return nil
}
