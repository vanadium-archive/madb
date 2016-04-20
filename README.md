# Madb: Multi-device Android Debug Bridge

[![Latest Release][release-image]][release-link]
[![Build Status][travis-image]][travis-link]
[![Coverage Status][coveralls-image]][coveralls-link]
[![API Documentation][godoc-image]][godoc-link]

Madb is a command line tool that wraps Android Debug Bridge (adb) and provides
various features for controlling multiple Android devices concurrently.

This tool is part of the Vanadium effort to build a framework and a set of
development tools to enable and ease the creation of multi-device user
interfaces and apps.

Madb releases are versioned according to
[Semantic Versioning 2.0.0](http://semver.org/spec/v2.0.0.html).

# Requirements

* OS
 - Linux (64bit)
 - Mac OS X (64bit)
 - Windows will be supported soon
* Tools
 - [ADB](http://developer.android.com/tools/help/adb.html): must be
installed and accessible from `PATH`. ADB comes with the Android SDK.
 - (Optional) [Flutter](https://flutter.io/): needed for using Flutter project
   specific features.

# Installing madb

## Downloading the Latest Release

The latest release of `madb` can be downloaded from the
[**releases page**](https://github.com/vanadium/madb/releases/latest).

Download the binary for your platform, and extract it to your desired location.
Make sure to add the location in your `PATH` environment variable to use `madb`
from anywhere.

## Using Go Get Command

If you have Go command line tool installed, you can also use the `go get`
command to get the most recent version of `madb`:

    $ go get github.com/vanadium/madb

This will install the tool under `<your GOPATH>/bin/madb`. To upgrade the tool
to the most recent version:

    $ go get -u github.com/vanadium/madb

**NOTE**: `go get` will always get the latest *development* version (i.e. the
current version in the `master` branch), not the stable release.

# Getting Started

This section introduces the most notable features of `madb`. To see the complete
list of features and their options, please use `madb help` or
`madb help <topic>`.

## Running an Android App on All Devices

If your have an Android application project using the
[Android plugin for Gradle](http://developer.android.com/tools/building/plugin-for-gradle.html)
as the build system (see if you have `build.gradle` file in your project
directory), you can type the following command from the project directory to
build, install, and launch your app on all connected devices in parallel:

    [From the project directory]
    $ madb start

Compare this with the following equivalent `adb` command:

    [For each device]
    $ adb -s <device_serial> shell am start -n <app ID>/<activity name>
    ...

The `madb start` command is more convenient in a few major ways.

First, with a single `madb start` command, it launches the app on all devices
and emulators. With only `adb`, you would have to manually issue the same
command with different device serials, or use a loop construct in a shell
script to get the same behavior, for example.

Next, `madb start` does not require any additional parameters. Internally,
`madb` reads the Gradle build scripts and the `AndroidManifest.xml` files in
the current directory, and automatically extracts the application ID and main
activity name to install and launch the correct app on all devices.

In case your build script contains some
[APK split](http://tools.android.com/tech-docs/new-build-system/user-guide/apk-splits)
configurations, `madb` will install the best matching APK for each device,
depending on the device supported ABIs and the screen density.

Also, the `adb` console outputs from each device is prefixed with the name of
the device, line by line. The following example output from `madb start`
demonstrates this.

```
[MyTablet]      7404 KB/s (21315847 bytes in 2.811s)
[MyTablet]              pkg: /data/local/tmp/app-universal-debug.apk
[MyPhone]       6270 KB/s (21315847 bytes in 3.319s)
[MyPhone]               pkg: /data/local/tmp/app-universal-debug.apk
[MyPhone]       Success
[MyPhone]       Stopping: com.yourcompany.yourapp
[MyPhone]       Starting: Intent { cmp=com.yourcompany.yourapp/.YourMainActivity }
[MyTablet]      Success
[MyTablet]      Stopping: com.yourcompany.yourapp
[MyTablet]      Starting: Intent { cmp=com.yourcompany.yourapp/.YourMainActivity }
$ _
```

Here, the names `MyPhone` and `MyTablet` are the device nicknames given using
`madb name` command. If these nicknames are not set, `madb` will show the
corresponding device serial number instead. When something goes wrong with one
of the devices, then you can see which of the available devices generated the
error by reading the prefixed device name.

If you want to skip the build step and just restart the app on all devices, you
can use the `-build=false` flag:

    $ madb start -build=false

## Running Arbitrary ADB Commands on All Devices

You can execute any `adb` commands on all devices in parallel using `madb exec`
command. For example, imagine you want to see the current time from all devices
to determine if there clocks are significantly skewed. You can use the following
`madb` command to check this:

```
$ madb exec shell date
[MyTablet]      Fri Apr 15 14:08:07 PDT 2016
[MyPhone]       Fri Apr 15 14:08:03 PDT 2016
$ _
```

**NOTE**: Launching an interactive shell (i.e. `adb shell` without arguments) on
all devices is not supported, and the behavior of `madb exec shell` without any
arguments is undefined. Always provide a specific shell command to execute.

If you want to copy a configuration file on your local computer to all devices,
you can use `madb exec` to issue `adb push` command on all devices as following:

```
$ madb exec push your.config /sdcard/
[MyTablet]      92 KB/s (4087 bytes in 0.043s)
[MyPhone]       87 KB/s (4087 bytes in 0.045s)
$ _
```

Also, you can see the live interleaving
[logcat](http://developer.android.com/tools/help/logcat.html)
messages coming from all devices by:

    $ madb exec logcat

## Giving Nicknames to Devices

As shown in the above examples, you can give human-friendly nicknames to your
devices. Once you give a nickname to a device, that nickname is used as the
console output prefix instead of its device serial number. To give a nickname,
you can use `madb name set` command.

    $ madb name set <device_serial> <nickname>

You can see the serial numbers of your devices with `adb devices -l` command.

```
$ adb devices -l
List of devices attached
01023f5e2fd2acab       device usb:3-5.3 product:bullhead model:Nexus_5X device:bullhead
HT4BVWV00000           device usb:3-5.4.2 product:volantisg model:Nexus_9 device:flounder_lte

$ _
```

In this output, `01023f5e2fd2acab` and `HT4BVWV00000` are the device serials.
To give `MyPhone` as the nickname of the Nexus 5X phone with serial
`01023f5e2fd2acab`,

    $ madb name set 01023f5e2fd2acab MyPhone

Similarly, you can give `MyTablet` as the nickname of the second device.

    $ madb name set HT4BVWV00000 MyTablet

To unset a given nickname, you can use `madb name unset`.

    $ madb name unset <device serial OR nickname>

For the purpose of displaying the nicknames as the console output prefixes, the
mapping between device serials and nicknames is kept 1:1. That is, there can be
only one nickname for any given device serial, and a nickname always resolves to
a single device serial.

## Specifying Devices

There can be situations where a certain `adb` command should be run only on a
subset of all available devices. `madb` provides a few flags for these
situations.

* `-d`: Restrict the command to only run on real devices.
* `-e`: Restrict the command to only run on emulators.
* `-n=<device1,device2,...>`:  Comma-separated device serials, qualifiers,
device indices (e.g., `@1`, `@2`), or nicknames (set by `madb name`). Command
will run only on specified devices.

For example, to launch your app only on emulators:

    $ madb -e start

To see the logcat messages from `Alice` and `Bob` devices and not from others:

    $ madb -n=Alice,Bob exec logcat

[coveralls-image]: https://img.shields.io/coveralls/vanadium/madb/master.svg?maxAge=2592000?style=flat-square
[coveralls-link]: https://coveralls.io/github/vanadium/madb?branch=master
[godoc-image]: https://godoc.org/github.com/vanadium/madb?status.svg
[godoc-link]: https://godoc.org/github.com/vanadium/madb
[release-image]: https://img.shields.io/github/release/vanadium/madb.svg?maxAge=2592000?style=flat-square
[release-link]: https://github.com/vanadium/madb/releases/latest
[travis-image]: https://img.shields.io/travis/vanadium/madb/master.svg?style=flat-square)
[travis-link]: https://travis-ci.org/vanadium/madb.svg?branch=master
