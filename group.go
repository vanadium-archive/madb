// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"v.io/x/lib/cmdline"
)

// TODO(youngseokyoon): use the groups for filtering devices.

var cmdMadbGroup = &cmdline.Command{
	Children: []*cmdline.Command{
		cmdMadbGroupAdd,
		cmdMadbGroupClearAll,
		cmdMadbGroupDelete,
		cmdMadbGroupList,
		cmdMadbGroupRemove,
		cmdMadbGroupRename,
	},
	Name:             "group",
	DontInheritFlags: true,
	Short:            "Manage device groups",
	Long: `
Manages device groups, each of which can have one or more device members. The
device groups can be used for specifying the target devices of other madb
commands.
`,
}

var cmdMadbGroupAdd = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupAdd, getDefaultConfigFilePath},
	Name:   "add",
	Short:  "Add members to a device group",
	Long: `
Adds members to a device group. This command also creates the group, if the
group does not exist yet. The device group can be used when specifying devices
in any madb commands.

When creating a new device group with this command, the provided name must not
conflict with an existing device nickname (see: madb help name set).

A group can contain another device group, in which case all the members of the
other group will also be considered as members of the current group.
`,
	ArgsName: "<group_name> <member1> [<member2> ...]",
	ArgsLong: `
<group_name> is an alpha-numeric string with no special characters or spaces.
This name must not be an existing device nickname.

<member> is a member specifier, which can be one of device serial, qualifier,
device index (e.g., '@1', '@2'), device nickname, or another device group.
`,
}

func runMadbGroupAdd(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) < 2 {
		return env.UsageErrorf("There must be at least two arguments.")
	}

	groupName := args[0]
	if !isValidName(groupName) {
		return fmt.Errorf("Not a valid group name: %q", groupName)
	}

	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}
	if isDeviceNickname(groupName, cfg) {
		return fmt.Errorf("The group name %q conflicts with a device nickname.", groupName)
	}

	members := removeDuplicates(args[1:])
	for _, member := range members {
		if err := isValidMember(member, cfg); err != nil {
			return fmt.Errorf("Invalid member %q: %v", member, err)
		}
	}

	oldMembers, ok := cfg.Groups[groupName]
	if !ok {
		oldMembers = []string{}
	}

	cfg.Groups[groupName] = removeDuplicates(append(oldMembers, members...))
	return writeConfig(cfg, filename)
}

var cmdMadbGroupClearAll = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupClearAll, getDefaultConfigFilePath},
	Name:   "clear-all",
	Short:  "Clear all the existing device groups",
	Long: `
Clears all the existing device groups.
`,
}

func runMadbGroupClearAll(env *cmdline.Env, args []string, filename string) error {
	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}

	// Reset the groups
	cfg.Groups = make(map[string][]string)

	return writeConfig(cfg, filename)
}

var cmdMadbGroupDelete = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupDelete, getDefaultConfigFilePath},
	Name:   "delete",
	Short:  "Delete an existing device group",
	Long: `
Deletes an existing device group.
`,
	ArgsName: "<group_name1> [<group_name2> ...]",
	ArgsLong: `
<group_name> the name of an existing device group.
You can specify more than one group names.
`,
}

func runMadbGroupDelete(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) < 1 {
		return env.UsageErrorf("There must be at least one argument.")
	}

	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}
	for _, groupName := range args {
		if !isGroupName(groupName, cfg) {
			return fmt.Errorf("Not an existing group name: %q", groupName)
		}
	}

	// Delete the groups
	for _, groupName := range args {
		delete(cfg.Groups, groupName)
	}

	return writeConfig(cfg, filename)
}

var cmdMadbGroupList = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupList, getDefaultConfigFilePath},
	Name:   "list",
	Short:  "List all the existing device groups",
	Long: `
Lists the name and members of all the existing device groups.
`,
}

func runMadbGroupList(env *cmdline.Env, args []string, filename string) error {
	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"Group Name", "Members"})
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetAutoFormatHeaders(false)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)

	data := make([][]string, 0, len(cfg.Groups))
	for group, members := range cfg.Groups {
		data = append(data, []string{group, strings.Join(members, " ")})
	}

	sort.Sort(byFirstElement(data))

	for _, row := range data {
		tw.Append(row)
	}
	tw.Render()

	return nil
}

var cmdMadbGroupRemove = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupRemove, getDefaultConfigFilePath},
	Name:   "remove",
	Short:  "Remove members from a device group",
	Long: `
Removes members from an existing device group. If there are no remaining members
after that, the group gets deleted.
`,
	ArgsName: "<group_name> <member1> [<member2> ...]",
	ArgsLong: `
<group_name> is an alpha-numeric string with no special characters or spaces.
This name must be an existing device group name.

<member> is a member specifier, which can be one of device serial, qualifier,
device index (e.g., '@1', '@2'), device nickname, or another device group.
`,
}

func runMadbGroupRemove(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) < 2 {
		return env.UsageErrorf("There must be at least two arguments.")
	}

	groupName := args[0]
	if !isValidName(groupName) {
		return fmt.Errorf("Not a valid group name: %q", groupName)
	}

	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}
	if !isGroupName(groupName, cfg) {
		return fmt.Errorf("Not an existing group name: %q", groupName)
	}

	members := removeDuplicates(args[1:])
	oldMembers := cfg.Groups[groupName]
	cfg.Groups[groupName] = subtractSlices(oldMembers, members)

	if len(cfg.Groups[groupName]) == 0 {
		delete(cfg.Groups, groupName)
	}

	return writeConfig(cfg, filename)
}

var cmdMadbGroupRename = &cmdline.Command{
	Runner: subCommandRunnerWithFilepath{runMadbGroupRename, getDefaultConfigFilePath},
	Name:   "rename",
	Short:  "Rename an existing device group",
	Long: `
Renames an existing device group.
`,
	ArgsName: "<old_name> <new_name>",
	ArgsLong: `
<old_name> is the name of an existing device group.

<new_name> is the new name for the existing group.
This must be an alpha-numeric string with no special characters or spaces,
and must not conflict with another existing device or group name.
`,
}

func runMadbGroupRename(env *cmdline.Env, args []string, filename string) error {
	// Check if the arguments are valid.
	if len(args) != 2 {
		return env.UsageErrorf("There must be exactly two arguments.")
	}

	oldName, newName := args[0], args[1]
	if !isValidName(oldName) {
		return fmt.Errorf("Not a valid group name: %q", oldName)
	}
	if !isValidName(newName) {
		return fmt.Errorf("Not a valid group name: %q", newName)
	}

	cfg, err := readConfig(filename)
	if err != nil {
		return err
	}
	if !isGroupName(oldName, cfg) {
		return fmt.Errorf("Not an existing group name: %q", oldName)
	}
	if isNameInUse(newName, cfg) {
		return fmt.Errorf("The provided name is already in use: %q", newName)
	}

	cfg.Groups[newName] = cfg.Groups[oldName]
	delete(cfg.Groups, oldName)

	return writeConfig(cfg, filename)
}

// isValidMember takes a member string given as an argument, and returns nil
// when the member string is valid. Otherwise, an error is returned inicating
// the reason why the given member string is not valid.
// TODO(youngseokyoon): reuse this function in madb.go.
func isValidMember(member string, cfg *config) error {
	if strings.HasPrefix(member, "@") {
		index, err := strconv.Atoi(member[1:])
		if err != nil || index <= 0 {
			return fmt.Errorf("Invalid device specifier %q. '@' sign must be followed by a numeric device index starting from 1.", member)
		}
		return nil
	} else if !isValidSerial(member) && !isValidName(member) {
		return fmt.Errorf("Invalid device specifier %q. Not a valid serial or a nickname.", member)
	}

	return nil
}

// removeDuplicates takes a string slice and removes all the duplicates.
func removeDuplicates(s []string) []string {
	result := make([]string, 0, len(s))

	used := map[string]bool{}
	for _, elem := range s {
		if !used[elem] {
			result = append(result, elem)
			used[elem] = true
		}
	}

	return result
}

// subtractSlices takes two slices and returns a new slice formed by removing
// all the elements in s2 from s1.
func subtractSlices(s1, s2 []string) []string {
	result := make([]string, 0, len(s1))

	m := map[string]bool{}
	for _, e2 := range s2 {
		m[e2] = true
	}

	for _, e1 := range s1 {
		if !m[e1] {
			result = append(result, e1)
		}
	}

	return result
}
