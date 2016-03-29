// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
)

func init() {
	initializePropertyCacheFlags(&cmdMadbInstall.Flags)
	initializeBuildFlags(&cmdMadbInstall.Flags)
}

var cmdMadbInstall = &cmdline.Command{
	Runner: subCommandRunner{initMadbInstall, runMadbInstallForDevice, true},
	Name:   "install",
	Short:  "Install your app on all devices",
	Long: `
Installs your app on all devices.

If the working directory contains a Gradle Android project (i.e., has "build.gradle"), this command
will first run a small Gradle script to extract the variant properties, which will be used to find
the best matching .apk for each device. These extracted properties are cached, and "madb install"
can be repeated without running this Gradle script again. The properties can be re-extracted by
clearing the cache by providing "-clear-cache" flag.

Once the variant properties are extracted, the best matching .apk for each device will be installed
in parallel.

This command is similar to running "gradlew :<moduleName>:<variantName>Install", but "madb install"
is more flexible: 1) you can install the app to a subset of the devices, and 2) the app is installed
concurrently, which saves a lot of time.


If the working directory contains a Flutter project (i.e., has "flutter.yaml"), this command will
run "flutter install --device-id <device serial>" for all devices.


To install your app for a specific user on a particular device, use 'madb user set' command to set
the default user ID for that device. (See 'madb help user' for more details.)

To install a specific .apk file to all devices, use "madb exec install <path_to_apk>" instead.
`,
}

func initMadbInstall(env *cmdline.Env, args []string, properties variantProperties) ([]string, error) {
	// If the "-build" flag is set, first run the relevant gradle tasks to build the .apk files
	// before installing the app to the devices.
	if isGradleProject(wd) && buildFlag {
		sh := gosh.NewShell(nil)
		defer sh.Cleanup()

		// Show the output from Gradle, so that users can see what's going on.
		sh.PropagateChildOutput = true
		sh.ContinueOnError = true

		wrapper, err := findGradleWrapper(wd)
		if err != nil {
			return nil, err
		}

		// Build the project by running ":<module>:assemble<Variant>" task.
		cmdArgs := []string{"--daemon", properties.AssembleTask}
		cmd := sh.Cmd(wrapper, cmdArgs...)
		cmd.Run()

		if err = sh.Err; err != nil {
			return nil, fmt.Errorf("Failed to build the app: %v", err)
		}
	}

	return args, nil
}

func runMadbInstallForDevice(env *cmdline.Env, args []string, d device, properties variantProperties) error {
	// The user is executing "madb install" explicitly, and the installation should not be skipped.
	return installVariantToDevice(d, properties, true)
}

func installVariantToDevice(d device, properties variantProperties, forceInstall bool) error {
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

		// Determine whether the app should be installed on the given device.
		shouldInstall, err := shouldInstallVariant(d, properties, bestOutput, forceInstall)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not determine whether the app should be installed on device %q. Attempting to install...", d.displayName())
			shouldInstall = true
		}

		// Run the "adb install" command to perform the installation.
		if shouldInstall {
			cmdArgs := []string{"-s", d.Serial, "install", "-r"}
			if d.UserID != "" {
				cmdArgs = append(cmdArgs, "--user", d.UserID)
			}
			cmdArgs = append(cmdArgs, bestOutput.OutputFilePath)
			cmd := sh.Cmd("adb", cmdArgs...)
			return runGoshCommandForDevice(cmd, d, true)
		}

		fmt.Printf("Device %q has the most recent version of the app already. Skipping the installation.\n", d.displayName())
	}

	if isFlutterProject(wd) {
		cmdArgs := []string{"install", "--device-id", d.Serial}
		cmd := sh.Cmd("flutter", cmdArgs...)
		return runGoshCommandForDevice(cmd, d, false)
	}

	return fmt.Errorf("Could not find the target app to be installed. Try running 'madb install' from a Gradle or Flutter project directory.")
}

// shouldInstallVariant determines whether the app should be installed on the given device or not.
func shouldInstallVariant(d device, properties variantProperties, bestOutput *variantOutput, forceInstall bool) (bool, error) {
	if forceInstall {
		return true, nil
	}

	// Check if the app is installed on this device.
	installed, err := isInstalled(d, properties)
	if err != nil {
		return false, err
	}
	if !installed {
		return true, nil
	}

	// TODO(youngseokyoon): check the "lastUpdateTime" property of the installed app.
	// For now, assume the app is outdated, and just return true to install.
	return true, nil
}

// isInstalled determines whether the app variant is already installed on the given device.
func isInstalled(d device, properties variantProperties) (bool, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	sh.ContinueOnError = true

	// Run "adb shell pm list packages --user <user_id> <app_id>".
	cmdArgs := []string{"-s", d.Serial, "shell", "pm", "list", "packages"}
	if d.UserID != "" {
		cmdArgs = append(cmdArgs, "--user", d.UserID)
	}
	cmdArgs = append(cmdArgs, properties.AppID)
	cmd := sh.Cmd("adb", cmdArgs...)
	output := cmd.Stdout()

	if sh.Err != nil {
		return false, sh.Err
	}

	// If the app is installed, the output should be in the form "package:<app_id>".
	if strings.TrimSpace(output) == fmt.Sprintf("package:%v", properties.AppID) {
		return true, nil
	}

	return false, nil
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
