# Madb: Multi-device Android Debug Bridge

[![GoDoc](https://godoc.org/github.com/vanadium/madb?status.svg)](https://godoc.org/github.com/vanadium/madb)

Madb is a command line tool that wraps Android Debug Bridge (adb) and provides
various features for controlling multiple Android devices concurrently.

This tool is part of the Vanadium effort to build a framework and a set of
development tools to enable and ease the creation of multi-device user
interfaces and apps.

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

## Using Go Get Command

If you have Go command line tool installed, you can use the `go get`
command to get the most recent version of `madb`:

    go get github.com/vanadium/madb

This will install the tool under `<your first GOPATH>/bin/madb`. Make sure to
add `<your first GOPATH>/bin` to your `PATH` environment variable before using
`madb`. To upgrade the tool to the most recent version:

    go get -u github.com/vanadium/madb

## Other Ways to Install

We plan to release pre-compiled binaries for different platforms on the
[releases page](https://github.com/vanadium/madb/releases), and via Homebrew in
the near future.

# Getting Started

This section introduces the most notable features of `madb`. To see the complete
list of features and their options, please use `madb help` or
`madb help <sub command>`.

## Running an Android App on Multiple Devices

## Running an ADB Command on Multiple Devices

## Giving Nicknames to Devices

## Specifying Devices

## Getting Logcat Output from Multiple Devices
