// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

var (
	replaceFlag bool
)

func init() {
	initializePropertyCacheFlags(&cmdMadbInstall.Flags)
	cmdMadbInstall.Flags.BoolVar(&replaceFlag, "r", true, `Replace the existing application. Same effect as the '-r' flag of 'adb install' command.`)
}

var cmdMadbInstall = &cmdline.Command{
	Runner: subCommandRunner{nil, runMadbInstallForDevice, true},
	Name:   "install",
	Short:  "Install your app on all devices",
	Long: `
Installs your app on all devices.

To install your app for a specific user on a particular device, use 'madb user set' command to set
the default user ID for that device. (See 'madb help user' for more details.)

If the working directory contains a Gradle Android project (i.e., has "build.gradle"), this command
will run a small Gradle script to extract the variant properties, which will be used to find the
best matching .apk for each device.

In this case, the extracted properties are cached, so that "madb install" can be repeated without
even running the Gradle script again. The IDs can be re-extracted by clearing the cache by providing
"-clear-cache" flag.

This command is similar to running "gradlew :<moduleName>:<variantName>Install", but the gradle
command is limited in that 1) it always installs the app to all connected devices, and 2) it
installs the app on one device at a time sequentially.

To install a specific .apk file to all devices, use "madb exec install <path_to_apk>" instead.
`,
}

func runMadbInstallForDevice(env *cmdline.Env, args []string, d device, properties variantProperties) error {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true
	if isGradleProject(wd) {
		// Get the necessary device properties.
		deviceAbis, err := getSupportedAbisForDevice(d)
		if err != nil {
			return err
		}

		deviceDensity, err := getScreenDensityForDevice(d)
		if err != nil {
			return err
		}

		// Compute the best output based on the device properties and the .apk filters.
		bestOutput := computeBestOutput(properties.VariantOutputs, properties.AbiFilters, deviceDensity, deviceAbis)
		if bestOutput == nil {
			return fmt.Errorf("Could not find the matching .apk for device %q", d.displayName())
		}

		// Run the install command.
		cmdArgs := []string{"-s", d.Serial, "install"}
		if replaceFlag {
			cmdArgs = append(cmdArgs, "-r")
		}
		if d.UserID != "" {
			cmdArgs = append(cmdArgs, "--user", d.UserID)
		}
		cmdArgs = append(cmdArgs, bestOutput.OutputFilePath)
		cmd := sh.Cmd("adb", cmdArgs...)
		return runGoshCommandForDevice(cmd, d, true)
	}

	if isFlutterProject(wd) {
		cmdArgs := []string{"install", "--device-id", d.Serial}
		cmd := sh.Cmd("flutter", cmdArgs...)
		return runGoshCommandForDevice(cmd, d, false)
	}

	return fmt.Errorf("Could not find the target app to be installed. Try running 'madb install' from a Gradle or Flutter project directory.")
}

// getSupportedAbisForDevice returns all the abis supported by the given device.
func getSupportedAbisForDevice(d device) ([]string, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	cmd := sh.Cmd("adb", "-s", d.Serial, "shell", "am", "get-config")
	output := cmd.Stdout()

	if sh.Err != nil {
		return nil, sh.Err
	}

	return parseSupportedAbis(output)
}

// parseSupportedAbis takes the output of "adb shell am get-config" command, and extracts the
// supported abis.
func parseSupportedAbis(output string) ([]string, error) {
	prefix := "abi: "
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			abis := strings.Split(line[len(prefix):], ",")
			return abis, nil
		}
	}

	return nil, fmt.Errorf("Could not extract the abi list from the device configuration output.")
}

// getScreenDensityForDevice returns the numeric screen dpi value of the given device.
func getScreenDensityForDevice(d device) (int, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	cmd := sh.Cmd("adb", "-s", d.Serial, "shell", "getprop")
	output := cmd.Stdout()

	if sh.Err != nil {
		return 0, sh.Err
	}

	return parseScreenDensity(output)
}

// parseScreenDensity takes the output of "adb shell getprop" command, and extracts the screen
// density value.
func parseScreenDensity(output string) (int, error) {
	// Each line in the output has the following format: "[<property_key>]: [<property_value>]"
	// The property key for the screen density is "ro.sf.lcd_density".
	// Look for the pattern "[ro.sf.lcd_density]: [(<value>)]" with regexp, with the value part
	// as a subexpression.
	key := "ro.sf.lcd_density"
	pattern := fmt.Sprintf(`\[%v\]: \[(.*)\]`, key)
	exp := regexp.MustCompile(pattern)

	matches := exp.FindStringSubmatch(output)
	// matches[1] is the first subexpression, which contains the density value we need.
	if matches != nil && len(matches) == 2 {
		return strconv.Atoi(matches[1])
	}

	return 0, fmt.Errorf("Could not extract the screen density from the device properties output.")
}

// getDensityResourceName converts the given numeric density value into a resource name such as
// "ldpi", "mdpi", etc.
func getDensityResourceName(density int) string {
	// Predefined density names.
	densityMap := map[int]string{
		0:   "anydpi",
		120: "ldpi",
		160: "mdpi",
		213: "tvdpi",
		240: "hdpi",
		320: "xhdpi",
		480: "xxhdpi",
		640: "xxxhdpi",
	}

	if name, ok := densityMap[density]; ok {
		return name
	}

	// Otherwise, return density + "dpi". (e.g., 280 -> "280dpi")
	return fmt.Sprintf("%vdpi", density)
}

// computeBestOutput returns the pointer of the best matching output among the multiple variant
// outputs, given the device density and the abis supported by the device. The logic of this
// function is similar to that of SplitOutputMatcher.java in the Android platform tools.
func computeBestOutput(variantOutputs []variantOutput, variantAbiFilters []string, deviceDensity int, deviceAbis []string) *variantOutput {
	densityName := getDensityResourceName(deviceDensity)
	matches := map[*variantOutput]bool{}

VariantOutputLoop:
	for i, vo := range variantOutputs {

	FilterLoop:
		for _, filter := range vo.Filters {

			switch filter.FilterType {
			case "ABI":
				for _, supportedAbi := range deviceAbis {
					if filter.Identifier == supportedAbi {
						// This filter is satisfied. Check for the next filter.
						continue FilterLoop
					}
				}

				// If the abi filter is not in the device supported abi list,
				// this variant output is not compatible with the device.
				// Check the next variant output.
				continue VariantOutputLoop

			case "DENSITY":
				if filter.Identifier != densityName {
					continue VariantOutputLoop
				}
			}
		}

		matches[&variantOutputs[i]] = true
	}

	// Return nil, if there are no matching variant outputs.
	if len(matches) == 0 {
		return nil
	}

	// Find the matching variant output with the maximum version code.
	// Iterate "variantOutputs" slice instead of "matches" map, in order to tie-break the matches
	// with same version codes by the order they are provided. (earlier defined output wins)
	var result, cur *variantOutput
	for i := range variantOutputs {
		cur = &variantOutputs[i]
		// Consider only the matching outputs
		if _, ok := matches[cur]; !ok {
			continue
		}

		if result == nil || result.VersionCode < cur.VersionCode {
			result = cur
		}
	}

	return result
}
