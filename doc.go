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
   clear-data  Clear your app data from all devices
   exec        Run the provided adb command on all devices and emulators
               concurrently
   install     Install your app on all devices
   name        Manage device nicknames
   start       Launch your app on all devices
   stop        Stop your app on all devices
   uninstall   Uninstall your app from all devices
   user        Manage default user settings for each device
   help        Display help for commands or topics

The madb flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

The global flags are:
 -metadata=<just specify -metadata to activate>
   Displays metadata for the program and exits.
 -time=false
   Dump timing information to stderr before exiting the program.

Madb clear-data - Clear your app data from all devices

Clears your app data from all devices.

To specify which user's data should be cleared, use 'madb user set' command to
set the default user ID for that device. (See 'madb help user' for more
details.)

Usage:
   madb clear-data [flags] [<application_id>]

<application_id> is usually the package name where the activities are defined.
(See:
http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

If the application ID is not specified, madb automatically determines which app
to be cleared, based on the build scripts found in the current working
directory.

If the working directory contains a Gradle Android project (i.e., has
"build.gradle"), run a small Gradle script to extract the application ID. In
this case, the extracted ID is cached, so that "madb clear-data" can be repeated
without even running the Gradle script again. The ID can be re-extracted by
clearing the cache by providing "-clear-cache" flag.

The madb clear-data flags are:
 -clear-cache=false
   Clear the cache and re-extract the variant properties such as the application
   ID and the main activity name. Only takes effect when no arguments are
   provided.
 -module=
   Specify which application module to use, when the current directory is the
   top level Gradle project containing multiple sub-modules. When not specified,
   the first available application module is used. Only takes effect when no
   arguments are provided.
 -variant=
   Specify which build variant to use. When not specified, the first available
   build variant is used. Only takes effect when no arguments are provided.

 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb install - Install your app on all devices

Installs your app on all devices.

To install your app for a specific user on a particular device, use 'madb user
set' command to set the default user ID for that device. (See 'madb help user'
for more details.)

If the working directory contains a Gradle Android project (i.e., has
"build.gradle"), this command will run a small Gradle script to extract the
variant properties, which will be used to find the best matching .apk for each
device.

In this case, the extracted properties are cached, so that "madb install" can be
repeated without even running the Gradle script again. The IDs can be
re-extracted by clearing the cache by providing "-clear-cache" flag.

This command is similar to running "gradlew :<moduleName>:<variantName>Install",
but the gradle command is limited in that 1) it always installs the app to all
connected devices, and 2) it installs the app on one device at a time
sequentially.

To install a specific .apk file to all devices, use "madb exec install
<path_to_apk>" instead.

Usage:
   madb install [flags]

The madb install flags are:
 -clear-cache=false
   Clear the cache and re-extract the variant properties such as the application
   ID and the main activity name. Only takes effect when no arguments are
   provided.
 -module=
   Specify which application module to use, when the current directory is the
   top level Gradle project containing multiple sub-modules. When not specified,
   the first available application module is used. Only takes effect when no
   arguments are provided.
 -r=true
   Replace the existing application. Same effect as the '-r' flag of 'adb
   install' command.
 -variant=
   Specify which build variant to use. When not specified, the first available
   build variant is used. Only takes effect when no arguments are provided.

 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb name set

Sets a human-friendly nickname that can be used when specifying the device in
any madb commands.

The device serial can be obtained using the 'adb devices -l' command. For
example, consider the following example output:

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
device qualifier (e.g., 'usb:3-3.4.2') obtained from 'adb devices -l' command
<nickname> is an alpha-numeric string with no special characters or spaces.

The madb name set flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb start - Launch your app on all devices

Launches your app on all devices.

To run your app as a specific user on a particular device, use 'madb user set'
command to set the default user ID for that device. (See 'madb help user' for
more details.)

Usage:
   madb start [flags] [<application_id> <activity_name>]

<application_id> is usually the package name where the activities are defined.
(See:
http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

<activity_name> is the Java class name for the activity you want to launch. If
the package name of the activity is different from the application ID, the
activity name must be a fully-qualified name (e.g.,
com.yourcompany.yourapp.MainActivity).

If either <application_id> or <activity_name> is provided, the other must be
provided as well.

If no arguments are specified, madb automatically determines which app to
launch, based on the build scripts found in the current working directory.

1) If the working directory contains a Flutter project (i.e., has
"flutter.yaml"), this command will run "flutter start --device-id <device
serial>" for all the specified devices.

2) If the working directory contains a Gradle Android project (i.e., has
"build.gradle"), this command will run a small Gradle script to extract the
application ID and the main activity name. In this case, the extracted IDs are
cached, so that "madb start" can be repeated without even running the Gradle
script again. The IDs can be re-extracted by clearing the cache by providing
"-clear-cache" flag.

The madb start flags are:
 -clear-cache=false
   Clear the cache and re-extract the variant properties such as the application
   ID and the main activity name. Only takes effect when no arguments are
   provided.
 -force-stop=true
   Force stop the target app before starting the activity.
 -module=
   Specify which application module to use, when the current directory is the
   top level Gradle project containing multiple sub-modules. When not specified,
   the first available application module is used. Only takes effect when no
   arguments are provided.
 -variant=
   Specify which build variant to use. When not specified, the first available
   build variant is used. Only takes effect when no arguments are provided.

 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb stop - Stop your app on all devices

Stops your app on all devices.

To stop your app for a specific user on a particular device, use 'madb user set'
command to set the default user ID for that device. (See 'madb help user' for
more details.)

Usage:
   madb stop [flags] [<application_id>]

<application_id> is usually the package name where the activities are defined.
(See:
http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

If the application ID is not specified, madb automatically determines which app
to stop, based on the build scripts found in the current working directory.

1) If the working directory contains a Flutter project (i.e., has
"flutter.yaml"), this command will run "flutter stop --device-id <device
serial>" for all the specified devices.

2) If the working directory contains a Gradle Android project (i.e., has
"build.gradle"), run a small Gradle script to extract the application ID. In
this case, the extracted ID is cached, so that "madb stop" can be repeated
without even running the Gradle script again. The ID can be re-extracted by
clearing the cache by providing "-clear-cache" flag.

The madb stop flags are:
 -clear-cache=false
   Clear the cache and re-extract the variant properties such as the application
   ID and the main activity name. Only takes effect when no arguments are
   provided.
 -module=
   Specify which application module to use, when the current directory is the
   top level Gradle project containing multiple sub-modules. When not specified,
   the first available application module is used. Only takes effect when no
   arguments are provided.
 -variant=
   Specify which build variant to use. When not specified, the first available
   build variant is used. Only takes effect when no arguments are provided.

 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb uninstall - Uninstall your app from all devices

Uninstall your app from all devices.

To uninstall your app for a specific user on a particular device, use 'madb user
set' command to set the default user ID for that device. (See 'madb help user'
for more details.)

Usage:
   madb uninstall [flags] [<application_id>]

<application_id> is usually the package name where the activities are defined.
(See:
http://tools.android.com/tech-docs/new-build-system/applicationid-vs-packagename)

If the application_id is not specified, madb automatically determines which app
to uninstall, based on the build scripts found in the current working directory.

If the working directory contains a Gradle Android project (i.e., has
"build.gradle"), run a small Gradle script to extract the application ID. In
this case, the extracted ID is cached, so that "madb uninstall" can be repeated
without even running the Gradle script again. The ID can be re-extracted by
clearing the cache by providing "-clear-cache" flag.

The madb uninstall flags are:
 -clear-cache=false
   Clear the cache and re-extract the variant properties such as the application
   ID and the main activity name. Only takes effect when no arguments are
   provided.
 -keep-data=false
   Keep the application data and cache directories. Equivalent to '-k' flag in
   'adb uninstall' command.
 -module=
   Specify which application module to use, when the current directory is the
   top level Gradle project containing multiple sub-modules. When not specified,
   the first available application module is used. Only takes effect when no
   arguments are provided.
 -variant=
   Specify which build variant to use. When not specified, the first available
   build variant is used. Only takes effect when no arguments are provided.

 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb user - Manage default user settings for each device

Manages default user settings for each device.

An Android device can have multiple user accounts, and each user account has a
numeric ID associated with it. Certain adb commands accept '--user <user_id>' as
a parameter to allow specifying which of the Android user account should be used
when running the command. The default behavior when the user ID is not provided
varies by the adb command being run.

Some madb commands internally run these adb commands which accept the '--user'
flag. You can let madb use different user IDs for different devices by storing
the default user ID for each device using 'madb user set' command. If the
default user ID is not set for a particular device, madb will not provide the
'--user' flag to the underlying adb command, and the current user will be used
for that device as a result.

Below is the list of madb commands which are affected by the default user ID
settings:

    madb clear-data
    madb start
    madb stop
    madb uninstall

For more details on how to obtain the user ID from an Android device, see 'madb
user help set'.

NOTE: Device specifier flags (-d, -e, -n) are ignored in all 'madb name'
commands.

Usage:
   madb user [flags] <command>

The madb user commands are:
   set         Set a default user ID to be used for the given device.
   unset       Unset the default user ID set by the 'madb user set' command.
   list        List all the existing default user IDs.
   clear-all   Clear all the existing default user settings.

The madb user flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb user set

Sets a default user ID to be used for the specified device, when there are
multiple user accounts on a single device.

The user IDs can be obtained using the 'adb [<device_serial>] shell pm list
users' command. Alternatively, you can use 'madb exec' if you want to specify
the device with a nickname. For example, running the following command:

    madb -n=MyPhone exec shell pm list users

will list the available users and their IDs on the MyPhone device. Consider the
following example output:

    [MyPhone]       Users:
    [MyPhone]               UserInfo{0:John Doe:13} running
    [MyPhone]               UserInfo{10:Work profile:30} running

There are two available users, "John Doe" and "Work profile". Each user is
assigned a "user ID", which appears on the left of the user name. In this case,
the user ID of "John Doe" is "0", and the user ID of the "Work profile" is "10".

To use the "Work profile" as the default user when running madb commands on this
device, run the following command:

    madb user set MyPhone 10

and then madb will use "Work profile" as the default user for device "MyPhone"
in any of the subsequence madb commands.

Usage:
   madb user set [flags] <device_serial> <user_id>

<device_serial> is the unique serial number for the device, which can be
obtained from 'adb devices'. <user_id> is one of the user IDs obtained from 'adb
shell pm list users' command.

The madb user set flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb user unset

Unsets the default user ID assigned by the 'madb user set' command for the
specified device.

Running this command without any device specifiers will unset the default users
only for the currently available devices and emulators, while keeping the
default user IDs for the other devices.

Usage:
   madb user unset [flags] <device_serial>

<device_serial> is the unique serial number for the device, which can be
obtained from 'adb devices'.

The madb user unset flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb user list

Lists all the currently stored default user IDs for devices.

Usage:
   madb user list [flags]

The madb user list flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

Madb user clear-all

Clears all the currently stored default user IDs for devices.

This command clears the default user IDs regardless of whether the device is
currently connected or not.

Usage:
   madb user clear-all [flags]

The madb user clear-all flags are:
 -d=false
   Restrict the command to only run on real devices.
 -e=false
   Restrict the command to only run on emulators.
 -n=
   Comma-separated device serials, qualifiers, device indices (e.g., '@1',
   '@2'), or nicknames (set by 'madb name'). A device index is specified by an
   '@' sign followed by the index of the device in the output of 'adb devices'
   command, starting from 1. Command will be run only on specified devices.

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
