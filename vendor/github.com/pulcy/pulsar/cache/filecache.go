// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"crypto/sha512"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/juju/errgo"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	cacheDir = "~/cache/pulcy"
)

var (
	maskAny    = errgo.MaskFunc(errgo.Any)
	cacheMutex sync.Mutex
)

func Clear(key string) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	dir, err := dir(key)
	if err != nil {
		return maskAny(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return maskAny(err)
	}
	return nil
}

func ClearAll() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	dir, err := rootDir()
	if err != nil {
		return maskAny(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return maskAny(err)
	}
	return nil
}

// Dir returns the cache directory for a given key.
// Returns: path, isValid, error
func Dir(key string, cacheValid time.Duration) (string, bool, error) {
	cachedir, err := dir(key)
	if err != nil {
		return "", false, maskAny(err)
	}

	// Lock
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Check if cache directory exists
	s, err := os.Stat(cachedir)
	isValid := false
	if err == nil {
		// Package cache directory exists, check age.
		if cacheValid != 0 && s.ModTime().Add(cacheValid).Before(time.Now()) {
			// Cache has become invalid
			if err := os.RemoveAll(cachedir); err != nil {
				return "", false, maskAny(err)
			}
		} else {
			// Cache is still valid
			isValid = true
		}
	} else {
		// cache directory not found, create needed
		isValid = false
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cachedir, 0777); err != nil {
		return "", false, maskAny(err)
	}

	return cachedir, isValid, nil
}

// dir returns the cache directory for a given key.
// Returns: path, error
func dir(key string) (string, error) {
	cachedirRoot, err := rootDir()
	if err != nil {
		return "", maskAny(err)
	}

	// Create hash of key
	hashBytes := sha512.Sum512([]byte(key))
	hash := fmt.Sprintf("%x", hashBytes)
	cachedir := filepath.Join(cachedirRoot, hash)

	return cachedir, nil
}

func rootDir() (string, error) {
	cachedirRoot, err := homedir.Expand(cacheDir)
	if err != nil {
		return "", maskAny(err)
	}

	return cachedirRoot, nil
}
