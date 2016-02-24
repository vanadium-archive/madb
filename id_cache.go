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

// projectIds contains the application ID and the activity name extracted from the Gradle scripts.
type projectIds struct {
	AppID, Activity string
}

// idCache is a map used for caching the ids extracted from the Gradle scripts, so that apps can be
// launched more quickly without running Gradle tasks.
type idCache map[variantKey]projectIds

func getIDCache(cacheFile string) (idCache, error) {
	return readIDCacheMap(cacheFile)
}

// Clears the cache entry from the given cacheFile.
func clearIDCacheEntry(key variantKey, cacheFile string) error {
	cache, err := getIDCache(cacheFile)
	if err != nil {
		return err
	}

	delete(cache, key)
	return writeIDCacheMap(cache, cacheFile)
}

// Adds a new entry in the id cache located at cacheFile and save the cache back to the file.
func writeIDCacheEntry(key variantKey, ids projectIds, cacheFile string) error {
	cache, err := getIDCache(cacheFile)
	if err != nil {
		return err
	}

	cache[key] = ids
	return writeIDCacheMap(cache, cacheFile)
}

// Reads the id cache map from the given file using gob-encoding.
func readIDCacheMap(filename string) (idCache, error) {
	f, err := os.Open(filename)
	if err != nil {
		// If the file does not exist, return an empty map without an error.
		if os.IsNotExist(err) {
			return idCache{}, nil
		}

		// An unexpected error occurred and should be returned.
		return nil, err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	result := idCache{}

	// Decoding might fail when the cache file is somehow corrupted, or when the cache schema is
	// updated.  In such cases, move on after resetting the cache file instead of exiting the app.
	if err := decoder.Decode(&result); err != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Could not decode the id cache file.  Resetting the cache.")
		if err := os.Remove(f.Name()); err != nil {
			return nil, err
		}

		return idCache{}, nil
	}

	return result, nil
}

// Writes the id cache map to the given file using gob-encoding.
func writeIDCacheMap(cache idCache, filename string) error {
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
