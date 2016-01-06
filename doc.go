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
   exec        Run the provided adb command on all the specified devices
               concurrently
   help        Display help for commands or topics

The global flags are:
 -metadata=<just specify -metadata to activate>
   Displays metadata for the program and exits.
 -time=false
   Dump timing information to stderr before exiting the program.

Madb exec - Run the provided adb command on all the specified devices concurrently

Runs the provided adb command on all the specified devices concurrently.

For example, the following line:

    madb -a exec push ./foo.txt /sdcard/foo.txt

copies the ./foo.txt file to /sdcard/foo.txt for all the currently connected
Android devices (specified by -a flag).

To see the list of available adb commands, type 'adb help'.

Usage:
   madb exec [flags] <command>

<command> is a normal adb command, which will be executed on all the specified
devices.

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
