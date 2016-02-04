// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated via go generate.
// DO NOT UPDATE MANUALLY

/*
Multi-device Android Debug Bridge

The madb command wraps Android Debug Bridge (adb) command line tool and provides
various features for controlling multiple Android devices concurrently.

Usage:
   madb [flags] <command>

The madb commands are:
   exec        Run the provided adb command on all devices and emulators
               concurrently
   name        Manage device nicknames
   help        Display help for commands or topics

The madb flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

The global flags are:
 -metadata=<just specify -metadata to activate>
   Displays metadata for the program and exits.
 -time=false
   Dump timing information to stderr before exiting the program.

Madb exec - Run the provided adb command on all devices and emulators concurrently

Runs the provided adb command on all devices and emulators concurrently.

For example, the following line:

    madb -a exec push ./foo.txt /sdcard/foo.txt

copies the ./foo.txt file to /sdcard/foo.txt for all the currently connected
Android devices.

To see the list of available adb commands, type 'adb help'.

Usage:
   madb exec [flags] <command>

<command> is a normal adb command, which will be executed on all devices and
emulators.

The madb exec flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb name - Manage device nicknames

Manages device nicknames, which are meant to be more human-friendly compared to
the device serials provided by adb tool.

NOTE: Device specifier flags (-d, -e, -n) are ignored in all 'madb name'
commands.

Usage:
   madb name [flags] <command>

The madb name commands are:
   set         Set a nickname to be used in place of the device serial.
   unset       Unset a nickname set by the 'madb name set' command.
   list        List all the existing nicknames.
   clear-all   Clear all the existing nicknames.

The madb name flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb name set

Sets a human-friendly nickname that can be used when specifying the device in
any madb commands.

The device serial can be obtain using the 'adb devices -l' command. For example,
consider the following example output:

    HT4BVWV00023           device usb:3-3.4.2 product:volantisg model:Nexus_9 device:flounder_lte

The first value, 'HT4BVWV00023', is the device serial. To assign a nickname for
this device, run the following command:

    madb name set HT4BVWV00023 MyTablet

and it will assign the 'MyTablet' nickname to the device serial 'HT4BVWV00023'.
The alternative device specifiers (e.g., 'usb:3-3.4.2', 'product:volantisg') can
also have nicknames.

When a nickname is set for a device serial, the nickname can be used to specify
the device within madb commands.

There can only be one nickname for a device serial. When the 'madb name set'
command is invoked with a device serial with an already assigned nickname, the
old one will be replaced with the newly provided one.

Usage:
   madb name set [flags] <device_serial> <nickname>

<device_serial> is a device serial (e.g., 'HT4BVWV00023') or an alternative
device specifier (e.g., 'usb:3-3.4.2') obtained from 'adb devices -l' command
<nickname> is an alpha-numeric string with no special characters or spaces.

The madb name set flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb name unset

Unsets a nickname assigned by the 'madb name set' command. Either the device
serial or the assigned nickname can be specified to remove the mapping.

Usage:
   madb name unset [flags] <device_serial | nickname>

There should be only one argument, which is either the device serial or the
nickname.

The madb name unset flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb name list

Lists all the currently stored nicknames of device serials.

Usage:
   madb name list [flags]

The madb name list flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb name clear-all

Clears all the currently stored nicknames of device serials.

Usage:
   madb name clear-all [flags]

The madb name clear-all flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, or nicknames (set by 'madb
   name').  Command will be run only on specified devices.

Madb help - Display help for commands or topics

Help with no args displays the usage of the parent command.

Help with args displays the usage of the specified sub-command or help topic.

"help ..." recursively displays help for all commands and topics.

Usage:
   madb help [flags] [command/topic ...]

[command/topic ...] optionally identifies a specific sub-command or help topic.

The madb help flags are:
 -style=compact
   The formatting style for help output:
      compact   - Good for compact cmdline output.
      full      - Good for cmdline output, shows all global flags.
      godoc     - Good for godoc processing.
      shortonly - Only output short description.
   Override the default by setting the CMDLINE_STYLE environment variable.
 -width=<terminal width>
   Format output to this target width in runes, or unlimited if width < 0.
   Defaults to the terminal width if available.  Override the default by setting
   the CMDLINE_WIDTH environment variable.
*/
package main
