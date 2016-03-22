// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// variantKey specifies a build variant in an Android Gradle project.
type variantKey struct {
	// Dir indicates the project directory where "build.gradle" resides.
	Dir string
	// Module indicates the name of the sub-module. Can be an empty string.
	Module string
	// Variant is the name of the build variant in an Android application module.
	// When there are no product flavors, there are only two build variants: "debug" and "release".
	Variant string
}

// propertyCache is a map used for caching the variant properties extracted from the Gradle scripts,
// so that apps can be launched more quickly without running Gradle tasks.
type propertyCache map[variantKey]variantProperties

func getPropertyCache(cacheFile string) (propertyCache, error) {
	return readPropertyCacheMap(cacheFile)
}

// Clears the cache entry from the given cacheFile.
func clearPropertyCacheEntry(key variantKey, cacheFile string) error {
	cache, err := getPropertyCache(cacheFile)
	if err != nil {
		return err
	}

	delete(cache, key)
	return writePropertyCacheMap(cache, cacheFile)
}

// Adds a new entry in the property cache located at cacheFile and save the cache back to the file.
func writePropertyCacheEntry(key variantKey, props variantProperties, cacheFile string) error {
	cache, err := getPropertyCache(cacheFile)
	if err != nil {
		return err
	}

	cache[key] = props
	return writePropertyCacheMap(cache, cacheFile)
}

// Reads the property cache map from the given file using gob-encoding.
func readPropertyCacheMap(filename string) (propertyCache, error) {
	f, err := os.Open(filename)
	if err != nil {
		// If the file does not exist, return an empty map without an error.
		if os.IsNotExist(err) {
			return propertyCache{}, nil
		}

		// An unexpected error occurred and should be returned.
		return nil, err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	result := propertyCache{}

	// Decoding might fail when the cache file is somehow corrupted, or when the cache schema is
	// updated. In such cases, move on after resetting the cache file instead of exiting the app.
	if err := decoder.Decode(&result); err != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Could not decode the property cache file. Resetting the cache.")
		if err := os.Remove(f.Name()); err != nil {
			return nil, err
		}

		return propertyCache{}, nil
	}

	return result, nil
}

// Writes the property cache map to the given file using gob-encoding.
func writePropertyCacheMap(cache propertyCache, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	return encoder.Encode(cache)
}

func getDefaultCacheFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "id_cache"), nil
}
