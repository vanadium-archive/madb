// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following enables go generate to generate the doc.go file.
//go:generate go run $JIRI_ROOT/release/go/src/v.io/x/lib/cmdline/testdata/gendoc.go .
//go:generate go run testdata/embed_gradle_script.go madb_init.gradle embedded_gradle.go gradleInitScript

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
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

	buildFlag bool

	wd string // working directory
)

func init() {
	cmdMadb.Flags.BoolVar(&allDevicesFlag, "d", false, `Restrict the command to only run on real devices.`)
	cmdMadb.Flags.BoolVar(&allEmulatorsFlag, "e", false, `Restrict the command to only run on emulators.`)
	cmdMadb.Flags.StringVar(&devicesFlag, "n", "", `Comma-separated device serials, qualifiers, device indices (e.g., '@1', '@2'), or nicknames (set by 'madb name'). A device index is specified by an '@' sign followed by the index of the device in the output of 'adb devices' command, starting from 1. Command will be run only on specified devices.`)

	// Store the current working directory.
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

// initializePropertyCacheFlags sets up the flags related to extracting and caching project properties.
func initializePropertyCacheFlags(flags *flag.FlagSet) {
	flags.BoolVar(&clearCacheFlag, "clear-cache", false, `Clear the cache and re-extract the variant properties such as the application ID and the main activity name. Only takes effect when no arguments are provided.`)
	flags.StringVar(&moduleFlag, "module", "", `Specify which application module to use, when the current directory is the top level Gradle project containing multiple sub-modules. When not specified, the first available application module is used. Only takes effect when no arguments are provided.`)
	flags.StringVar(&variantFlag, "variant", "", `Specify which build variant to use. When not specified, the first available build variant is used. Only takes effect when no arguments are provided.`)
}

// initializeBuildFlags sets up the flags related to running Gradle build tasks.
func initializeBuildFlags(flags *flag.FlagSet) {
	flags.BoolVar(&buildFlag, "build", true, `Build the target app variant before installing or running the app.`)
}

var cmdMadb = &cmdline.Command{
	Children: []*cmdline.Command{
		cmdMadbClearData,
		cmdMadbExec,
		cmdMadbInstall,
		cmdMadbName,
		cmdMadbStart,
		cmdMadbStop,
		cmdMadbUninstall,
		cmdMadbUser,
	},
	Name:  "madb",
	Short: "Multi-device Android Debug Bridge",
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
	Index      int
	UserID     string
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
func getDevices(nicknameFile string, userFile string) ([]device, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	output := sh.Cmd("adb", "devices", "-l").Stdout()

	nicknameSerialMap, err := readMapFromFile(nicknameFile)
	if err != nil {
		return nil, err
	}

	serialUserMap, err := readMapFromFile(userFile)
	if err != nil {
		return nil, err
	}

	return parseDevicesOutput(output, nicknameSerialMap, serialUserMap)
}

// Parses the output generated from "adb devices -l" command and return the list of device serial numbers
// Devices that are currently offline are excluded from the returned list.
func parseDevicesOutput(output string, nicknameSerialMap map[string]string, serialUserMap map[string]string) ([]device, error) {
	lines := strings.Split(output, "\n")

	result := []device{}

	// Check the first line of the output
	if len(lines) <= 0 || strings.TrimSpace(lines[0]) != "List of devices attached" {
		return result, fmt.Errorf("The output from 'adb devices -l' command does not look as expected.")
	}

	// Iterate over all the device serial numbers, starting from the second line.
	for i, line := range lines[1:] {
		fields := strings.Fields(line)

		if len(fields) <= 1 || fields[1] == "offline" {
			continue
		}

		// Fill in the device serial, all the qualifiers, and the device index.
		d := device{
			Serial:     fields[0],
			Qualifiers: fields[2:],
			Index:      i + 1,
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
		for nickname, serial := range nicknameSerialMap {
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

		// Determine whether there is a default user ID set by 'madb user'.
		if userID, ok := serialUserMap[d.Serial]; ok {
			d.UserID = userID
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

	userFile, err := getDefaultUserFilePath()
	if err != nil {
		return nil, err
	}

	allDevices, err := getDevices(nicknameFile, userFile)
	if err != nil {
		return nil, err
	}

	filtered, err := filterSpecifiedDevices(allDevices)
	if err != nil {
		return nil, err
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("No devices matching the device specifiers.")
	}

	return filtered, nil
}

type deviceSpec struct {
	index int
	token string
}

func filterSpecifiedDevices(devices []device) ([]device, error) {
	// If no device specifier flags are set, run on all devices and emulators.
	if noDevicesSpecified() {
		return devices, nil
	}

	result := make([]device, 0, len(devices))

	var specs = []deviceSpec{}
	if devicesFlag != "" {
		tokens := strings.Split(devicesFlag, ",")
		for _, token := range tokens {
			if strings.HasPrefix(token, "@") {
				index, err := strconv.Atoi(token[1:])
				if err != nil || index <= 0 {
					return nil, fmt.Errorf("Invalid device specifier %q. '@' sign must be followed by a numeric device index starting from 1.", token)
				}
				specs = append(specs, deviceSpec{index, ""})
			} else {
				specs = append(specs, deviceSpec{0, token})
			}
		}
	}

	for _, d := range devices {
		if shouldIncludeDevice(d, specs) {
			result = append(result, d)
		}
	}

	return result, nil
}

func noDevicesSpecified() bool {
	return allDevicesFlag == false &&
		allEmulatorsFlag == false &&
		devicesFlag == ""
}

func shouldIncludeDevice(d device, specs []deviceSpec) bool {
	if allDevicesFlag && d.Type == realDevice {
		return true
	}

	if allEmulatorsFlag && d.Type == emulator {
		return true
	}

	for _, spec := range specs {
		// Ignore empty tokens
		if spec.index == 0 && spec.token == "" {
			continue
		}

		if spec.index > 0 {
			if d.Index == spec.index {
				return true
			}
			continue
		}

		if d.Serial == spec.token || d.Nickname == spec.token {
			return true
		}

		for _, qualifier := range d.Qualifiers {
			if qualifier == spec.token {
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
	init func(env *cmdline.Env, args []string, properties variantProperties) ([]string, error)
	// subCmd defines the behavior of the sub command which will run on all the
	// devices in parallel.
	subCmd func(env *cmdline.Env, args []string, d device, properties variantProperties) error
	// extractProperties indicates whether this subCommand needs the extracted
	// project properties.
	extractProperties bool
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

	// Extract the properties if needed.
	properties := variantProperties{}
	if r.extractProperties && isGradleProject(wd) {
		properties, err = getProjectPropertiesUsingDefaultCache()
		if err != nil {
			return err
		}
	}

	// Run the init function when provided.
	if r.init != nil {
		newArgs, err := r.init(env, args, properties)
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
			if err := r.subCmd(env, args, deviceCopy, properties); err != nil {
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

func runGoshCommandForDevice(cmd *gosh.Cmd, d device, printUserID bool) error {
	var prefix string
	if printUserID && d.UserID != "" {
		prefix = "[" + d.displayName() + ":" + d.UserID + "]\t"
	} else {
		prefix = "[" + d.displayName() + "]\t"
	}

	stdout := textutil.PrefixLineWriter(os.Stdout, prefix)
	stderr := textutil.PrefixLineWriter(os.Stderr, prefix)
	cmd.AddStdoutWriter(stdout)
	cmd.AddStderrWriter(stderr)
	cmd.Run()
	stdout.Flush()
	stderr.Flush()

	return cmd.Shell().Err
}

func initMadbCommand(env *cmdline.Env, args []string, properties variantProperties, flutterPassthrough bool, activityNameRequired bool) ([]string, error) {
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
		args = []string{properties.AppID, properties.Activity}[:numRequiredArgs]
	}

	return args, nil
}

func getProjectPropertiesUsingDefaultCache() (variantProperties, error) {
	cacheFile, err := getDefaultCacheFilePath()
	if err != nil {
		return variantProperties{}, err
	}

	key := variantKey{wd, moduleFlag, variantFlag}
	return getProjectProperties(extractPropertiesFromGradle, key, clearCacheFlag, cacheFile)
}

type variantProperties struct {
	ProjectPath    string
	VariantName    string
	CleanTask      string
	AssembleTask   string
	AppID          string
	Activity       string
	AbiFilters     []string
	VariantOutputs []variantOutput
}

type variantOutput struct {
	Name           string
	OutputFilePath string
	VersionCode    int
	Filters        []filter
}

type filter struct {
	FilterType string
	Identifier string
}

type propertyExtractorFunc func(variantKey) (variantProperties, error)

// getProjectProperties returns the project properties for the given build variant.
// It returns the cached values when the variant is found in the cache file, unless the clearCache
// argument is true. Otherwise, it calls extractPropertiesFromGradle to extract those properties by
// running Gradle scripts.
func getProjectProperties(extractor propertyExtractorFunc, key variantKey, clearCache bool, cacheFile string) (variantProperties, error) {
	if clearCache {
		clearPropertyCacheEntry(key, cacheFile)
	} else {
		// See if the current configuration appears in the cache.
		cache, err := getPropertyCache(cacheFile)
		if err != nil {
			return variantProperties{}, err
		}

		if properties, ok := cache[key]; ok {
			fmt.Println("NOTE: Cached IDs are being used. Use '-clear-cache' flag to clear the cache and extract the IDs from Gradle scripts again.")
			return properties, nil
		}
	}

	fmt.Println("Running Gradle to extract the application ID and the main activity name...")
	properties, err := extractor(key)
	if err != nil {
		return variantProperties{}, err
	}

	// Write these properties to the cache.
	if err := writePropertyCacheEntry(key, properties, cacheFile); err != nil {
		return variantProperties{}, fmt.Errorf("Could not write properties to the cache file: %v", err)
	}

	return properties, nil
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

func extractPropertiesFromGradle(key variantKey) (variantProperties, error) {
	sh := gosh.NewShell(nil)
	defer sh.Cleanup()

	// Continue on error instead of panicking and check the sh.Err value afterwards.
	// Gradle build will finish with exit code other than 0, when it fails.
	// In such cases, we want to show users meaningful error messages instead of stacktraces.
	sh.PropagateChildOutput = true
	sh.ContinueOnError = true

	wrapper, err := findGradleWrapper(key.Dir)
	if err != nil {
		return variantProperties{}, err
	}

	// Write the init script in a temp file.
	initScript := sh.MakeTempFile()
	initScript.WriteString(gradleInitScript)
	initScript.Close()

	// Create a temporary file in which Gradle can write the results.
	outputFile := sh.MakeTempFile()

	// Run the gradle wrapper to extract the application ID and the main activity name from the build scripts.
	cmdArgs := []string{"--daemon", "-q", "-I", initScript.Name(), "-PmadbOutputFile=" + outputFile.Name()}

	// Specify the project directory. If the module name is explicitly set, combine it with the base directory.
	cmdArgs = append(cmdArgs, "-p", filepath.Join(key.Dir, key.Module))

	// Specify the variant
	if key.Variant != "" {
		cmdArgs = append(cmdArgs, "-PmadbVariant="+key.Variant)
	}

	// Specify the tasks
	cmdArgs = append(cmdArgs, "madbExtractVariantProperties")

	cmd := sh.Cmd(wrapper, cmdArgs...)
	cmd.Run()

	if err = sh.Err; err != nil {
		return variantProperties{}, err
	}

	// Read what is written in the temporary file.
	// The file must be in JSON format.
	result := variantProperties{}
	decoder := json.NewDecoder(outputFile)
	if err = decoder.Decode(&result); err != nil {
		return variantProperties{}, fmt.Errorf("Could not extract the application ID and the main activity name: %v", err)
	}

	return result, nil
}

// readMapFromFile reads the provided file and reconstructs the string => string map.
// When the file does not exist, it returns an empty map (instead of nil), so that callers can safely add new entries.
func readMapFromFile(filename string) (map[string]string, error) {
	result := make(map[string]string)

	// The file may not exist or be empty when there are no stored data.
	if stat, err := os.Stat(filename); os.IsNotExist(err) || (err == nil && stat.Size() == 0) {
		return result, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	// Decoding might fail when the file is somehow corrupted, or when the schema is updated.
	// In such cases, move on after resetting the cache file instead of exiting the app.
	if err := decoder.Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Could not decode the file: %q. Resetting the file.\n", err)
		if err := os.Remove(f.Name()); err != nil {
			return nil, err
		}

		return make(map[string]string), nil
	}

	return result, nil
}

// writeMapToFile takes a string => string map and writes it into the provided file name.
func writeMapToFile(data map[string]string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(data)
}

type pathProvider func() (string, error)

// subCommandRunnerWithFilepath is an adapter that turns the madb name/user subcommand functions into cmdline.Runners.
type subCommandRunnerWithFilepath struct {
	subCmd func(*cmdline.Env, []string, string) error
	pp     pathProvider
}

var _ cmdline.Runner = (*subCommandRunnerWithFilepath)(nil)

// Run implements the cmdline.Runner interface by providing the default name file path
// as the third string argument of the underlying run function.
func (f subCommandRunnerWithFilepath) Run(env *cmdline.Env, args []string) error {
	p, err := f.pp()
	if err != nil {
		return err
	}

	return f.subCmd(env, args, p)
}
