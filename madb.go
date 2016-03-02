// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $JIRI_ROOT/release/go/src/v.io/x/lib/cmdline/testdata/gendoc.go .

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"v.io/x/lib/cmdline"
	"v.io/x/lib/gosh"
	"v.io/x/lib/textutil"
)

var (
	allDevicesFlag   bool
	allEmulatorsFlag bool
	devicesFlag      string

	clearCacheFlag bool
	moduleFlag     string
	variantFlag    string

	wd string // working directory
)

func init() {
	cmdMadb.Flags.BoolVar(&allDevicesFlag, "d", false, `Restrict the command to only run on real devices.`)
	cmdMadb.Flags.BoolVar(&allEmulatorsFlag, "e", false, `Restrict the command to only run on emulators.`)
	cmdMadb.Flags.StringVar(&devicesFlag, "n", "", `Comma-separated device serials, qualifiers, or nicknames (set by 'madb name').  Command will be run only on specified devices.`)

	// Store the current working directory.
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

// initializes flags related to extracting and caching project ids.
func initializeIDCacheFlags(flags *flag.FlagSet) {
	flags.BoolVar(&clearCacheFlag, "clear-cache", false, `Clear the cache and re-extract the application ID and the main activity name.  Only takes effect when no arguments are provided.`)
	flags.StringVar(&moduleFlag, "module", "", `Specify which application module to use, when the current directory is the top level Gradle project containing multiple sub-modules.  When not specified, the first available application module is used.  Only takes effect when no arguments are provided.`)
	flags.StringVar(&variantFlag, "variant", "", `Specify which build variant to use.  When not specified, the first available build variant is used.  Only takes effect when no arguments are provided.`)
}

var cmdMadb = &cmdline.Command{
	Children: []*cmdline.Command{cmdMadbClearData, cmdMadbExec, cmdMadbName, cmdMadbStart, cmdMadbStop, cmdMadbUninstall},
	Name:     "madb",
	Short:    "Multi-device Android Debug Bridge",
	Long: `
Multi-device Android Debug Bridge

The madb command wraps Android Debug Bridge (adb) command line tool
and provides various features for controlling multiple Android devices concurrently.
`,
}

func main() {
	cmdline.Main(cmdMadb)
}

// Makes sure that adb server is running.
// Intended to be called at the beginning of each subcommand.
func startAdbServer() error {
	// TODO(youngseokyoon): search for installed adb tool more rigourously.
	if err := exec.Command("adb", "start-server").Run(); err != nil {
		return fmt.Errorf("Failed to start adb server. Please make sure that adb is in your PATH: %v", err)
	}

	return nil
}

type deviceType string

const (
	emulator   deviceType = "Emulator"
	realDevice deviceType = "RealDevice"
)

type device struct {
	Serial     string
	Type       deviceType
	Qualifiers []string
	Nickname   string
}

// Returns the display name which is intended to be used as the console output prefix.
// This would be the nickname of the device if there is one; otherwise, the serial number is used.
func (d device) displayName() string {
	if d.Nickname != "" {
		return d.Nickname
	}

	return d.Serial
}

// Runs "adb devices -l" command, and parses the result to get all the device serial numbers.
func getDevices(nicknameFile string) ([]device, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	output := sh.Cmd("adb", "devices", "-l").Stdout()

	nsm, err := readNicknameSerialMap(nicknameFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not read the nickname file.")
	}

	return parseDevicesOutput(output, nsm)
}

// Parses the output generated from "adb devices -l" command and return the list of device serial numbers
// Devices that are currently offline are excluded from the returned list.
func parseDevicesOutput(output string, nsm map[string]string) ([]device, error) {
	lines := strings.Split(output, "\n")

	result := []device{}

	// Check the first line of the output
	if len(lines) <= 0 || strings.TrimSpace(lines[0]) != "List of devices attached" {
		return result, fmt.Errorf("The output from 'adb devices -l' command does not look as expected.")
	}

	// Iterate over all the device serial numbers, starting from the second line.
	for _, line := range lines[1:] {
		fields := strings.Fields(line)

		if len(fields) <= 1 || fields[1] == "offline" {
			continue
		}

		// Fill in the device serial and all the qualifiers.
		d := device{
			Serial:     fields[0],
			Qualifiers: fields[2:],
		}

		// Determine whether this device is an emulator or a real device.
		if strings.HasPrefix(d.Serial, "emulator") {
			d.Type = emulator
		} else {
			d.Type = realDevice
		}

		// Determine whether there is a nickname defined for this device,
		// so that the console output prefix can display the nickname instead of the serial.
	NSMLoop:
		for nickname, serial := range nsm {
			if d.Serial == serial {
				d.Nickname = nickname
				break
			}

			for _, qualifier := range d.Qualifiers {
				if qualifier == serial {
					d.Nickname = nickname
					break NSMLoop
				}
			}
		}

		result = append(result, d)
	}

	return result, nil
}

// Gets all the devices specified by the device specifier flags.
// Intended to be used by most of the madb sub-commands except for 'madb name'.
func getSpecifiedDevices() ([]device, error) {
	nicknameFile, err := getDefaultNameFilePath()
	if err != nil {
		return nil, err
	}

	allDevices, err := getDevices(nicknameFile)
	if err != nil {
		return nil, err
	}

	filtered := filterSpecifiedDevices(allDevices)

	if len(filtered) == 0 {
		return nil, fmt.Errorf("No devices matching the device specifiers.")
	}

	return filtered, nil
}

func filterSpecifiedDevices(devices []device) []device {
	// If no device specifier flags are set, run on all devices and emulators.
	if noDevicesSpecified() {
		return devices
	}

	result := make([]device, 0, len(devices))

	for _, d := range devices {
		if shouldIncludeDevice(d) {
			result = append(result, d)
		}
	}

	return result
}

func noDevicesSpecified() bool {
	return allDevicesFlag == false &&
		allEmulatorsFlag == false &&
		devicesFlag == ""
}

func shouldIncludeDevice(d device) bool {
	if allDevicesFlag && d.Type == realDevice {
		return true
	}

	if allEmulatorsFlag && d.Type == emulator {
		return true
	}

	tokens := strings.Split(devicesFlag, ",")
	for _, token := range tokens {
		// Ignore empty tokens
		if token == "" {
			continue
		}

		if d.Serial == token || d.Nickname == token {
			return true
		}

		for _, qualifier := range d.Qualifiers {
			if qualifier == token {
				return true
			}
		}
	}

	return false
}

// Returns the config dir located at "~/.madb"
func getConfigDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("Could not find the HOME directory.")
	}

	configDir := filepath.Join(home, ".madb")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

type subCommandRunner struct {
	// init is an optional function that does some initial work that should only
	// be performed once, before directing the command to all the devices.
	// The returned string slice becomes the new set of arguments passed into
	// the sub command.
	init func(env *cmdline.Env, args []string) ([]string, error)
	// subCmd defines the behavior of the sub command which will run on all the
	// devices in parallel.
	subCmd func(env *cmdline.Env, args []string, d device) error
}

var _ cmdline.Runner = (*subCommandRunner)(nil)

// Invokes the sub command on all the devices in parallel.
func (r subCommandRunner) Run(env *cmdline.Env, args []string) error {
	if err := startAdbServer(); err != nil {
		return err
	}

	devices, err := getSpecifiedDevices()
	if err != nil {
		return err
	}

	// Run the init function when provided.
	if r.init != nil {
		newArgs, err := r.init(env, args)
		if err != nil {
			return err
		}

		args = newArgs
	}

	wg := sync.WaitGroup{}

	var errs []error
	var errDevices []device

	for _, d := range devices {
		// Capture the current value.
		deviceCopy := d

		wg.Add(1)
		go func() {
			if err := r.subCmd(env, args, deviceCopy); err != nil {
				errs = append(errs, err)
				errDevices = append(errDevices, deviceCopy)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	// Report any errors returned from the go-routines.
	if errs != nil {
		buffer := bytes.Buffer{}
		buffer.WriteString("Error occurred while running the command on the following devices:")
		for i := 0; i < len(errs); i++ {
			buffer.WriteString("\n[" + errDevices[i].displayName() + "]\t" + errs[i].Error())
		}
		return fmt.Errorf(buffer.String())
	}

	return nil
}

func runGoshCommandForDevice(cmd *gosh.Cmd, d device) error {
	stdout := textutil.PrefixLineWriter(os.Stdout, "["+d.displayName()+"]\t")
	stderr := textutil.PrefixLineWriter(os.Stderr, "["+d.displayName()+"]\t")
	cmd.AddStdoutWriter(stdout)
	cmd.AddStderrWriter(stderr)
	cmd.Run()
	stdout.Flush()
	stderr.Flush()

	return cmd.Shell().Err
}

func initMadbCommand(env *cmdline.Env, args []string, flutterPassthrough bool, activityNameRequired bool) ([]string, error) {
	var numRequiredArgs int
	var requiredArgsStr string

	if activityNameRequired {
		numRequiredArgs = 2
		requiredArgsStr = "two arguments"
	} else {
		numRequiredArgs = 1
		requiredArgsStr = "one argument"
	}

	// Pass the arguments through if all the required arguments are provided, or if it is a flutter project.
	if len(args) == numRequiredArgs || (flutterPassthrough && isFlutterProject(wd)) {
		return args, nil
	}

	if len(args) != 0 {
		return nil, fmt.Errorf("You mush provide either zero arguments or exactly %v.", requiredArgsStr)
	}

	// Try to extract the application ID and the main activity name from the Gradle scripts.
	if isGradleProject(wd) {
		cacheFile, err := getDefaultCacheFilePath()
		if err != nil {
			return nil, err
		}

		key := variantKey{wd, moduleFlag, variantFlag}
		ids, err := getProjectIds(extractIdsFromGradle, key, clearCacheFlag, cacheFile)
		if err != nil {
			return nil, err
		}

		args = []string{ids.AppID, ids.Activity}[:numRequiredArgs]
	}

	return args, nil
}

type idExtractorFunc func(variantKey) (projectIds, error)

// Returns the project ids for the given build variant.  It returns the cached values when the
// variant is found in the cache file, unless the clearCache argument is true.  Otherwise, it calls
// extractIdsFromGradle to extract those ids by running Gradle scripts.
func getProjectIds(extractor idExtractorFunc, key variantKey, clearCache bool, cacheFile string) (projectIds, error) {
	if clearCache {
		clearIDCacheEntry(key, cacheFile)
	} else {
		// See if the current configuration appears in the cache.
		cache, err := getIDCache(cacheFile)
		if err != nil {
			return projectIds{}, err
		}

		if ids, ok := cache[key]; ok {
			fmt.Println("NOTE: Cached IDs are being used.  Use '-clear-cache' flag to clear the cache and extract the IDs from Gradle scripts again.")
			return ids, nil
		}
	}

	fmt.Println("Running Gradle to extract the application ID and the main activity name...")
	ids, err := extractor(key)
	if err != nil {
		return projectIds{}, err
	}

	// Write these ids to the cache.
	if err := writeIDCacheEntry(key, ids, cacheFile); err != nil {
		return projectIds{}, fmt.Errorf("Could not write ids to the cache file: %v", err)
	}

	return ids, nil
}

func isFlutterProject(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "flutter.yaml"))
	return err == nil
}

func isGradleProject(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "build.gradle"))
	return err == nil
}

// Looks for the Gradle wrapper script file ("gradlew"), starting from the current directory.
func findGradleWrapper(dir string) (string, error) {
	curDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		wrapperPath := filepath.Join(curDir, "gradlew")

		// Return the path of the gradle wrapper script if it is found.
		_, err := os.Stat(wrapperPath)
		if err == nil {
			// Found Gradle wrapper. Return the absolute path.
			return wrapperPath, nil
		} else if !os.IsNotExist(err) {
			// This is an unexpected error and should be returned.
			return "", err
		}

		// Search again in the parent directory.
		parentDir := path.Dir(curDir)
		if curDir == parentDir || parentDir == "." {
			break
		}

		curDir = parentDir
	}

	return "", fmt.Errorf("Could not find the Gradle wrapper in dir %q or its parent directories.", dir)
}

// TODO(youngseokyoon): find a better way to distribute the gradle script.
func findGradleInitScript() (string, error) {
	jiriRoot := os.Getenv("JIRI_ROOT")
	if jiriRoot == "" {
		return "", fmt.Errorf("JIRI_ROOT environment variable is not set")
	}

	initScript := filepath.Join(jiriRoot, "release", "go", "src", "v.io", "x", "devtools", "madb", "madb_init.gradle")
	if _, err := os.Stat(initScript); err != nil {
		return "", err
	}

	return initScript, nil
}

func extractIdsFromGradle(key variantKey) (ids projectIds, err error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	// Continue on error instead of panicking and check the sh.Err value afterwards.
	// Gradle build will finish with exit code other than 0, when it fails.
	// In such cases, we want to show users meaningful error messages instead of stacktraces.
	sh.PropagateChildOutput = true
	sh.ContinueOnError = true

	wrapper, err := findGradleWrapper(key.Dir)
	if err != nil {
		return
	}

	initScript, err := findGradleInitScript()
	if err != nil {
		err = fmt.Errorf("Could not find the madb_init.gradle script: %v", err)
		return
	}

	// Create a temporary file in which Gradle can write the results.
	outputFile := sh.MakeTempFile()

	// Run the gradle wrapper to extract the application ID and the main activity name from the build scripts.
	cmdArgs := []string{"--daemon", "-q", "-I", initScript, "-PmadbOutputFile=" + outputFile.Name()}

	// Specify the project directory. If the module name is explicitly set, combine it with the base directory.
	cmdArgs = append(cmdArgs, "-p", filepath.Join(key.Dir, key.Module))

	// Specify the variant
	if key.Variant != "" {
		cmdArgs = append(cmdArgs, "-PmadbVariant="+key.Variant)
	}

	// Specify the tasks
	cmdArgs = append(cmdArgs, "madbExtractApplicationId", "madbExtractMainActivity")

	cmd := sh.Cmd(wrapper, cmdArgs...)
	cmd.Run()

	if err = sh.Err; err != nil {
		return
	}

	// Read what is written in the temporary file.
	var bytes []byte
	bytes, err = ioutil.ReadFile(outputFile.Name())
	if err != nil {
		return
	}

	lines := strings.Split(string(bytes[:]), "\n")
	if len(lines) != 3 {
		err = fmt.Errorf("Could not extract the application ID and the main activity name.")
		return
	}

	ids = projectIds{lines[0], lines[1]}
	return
}
