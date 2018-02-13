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

package golang

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/settings"
	"github.com/pulcy/pulsar/util"
)

type FlattenFlags struct {
	VendorDir string
	TargetDir string // Defaults to $GOPATH/src
	NoRemove  bool
}

// Flatten copies all directories found in the given vendor directory to the GOPATH
// and flattens all vendor directories found in the GOPATH.
func Flatten(log *log.Logger, flags *FlattenFlags) error {
	vendorDir, err := filepath.Abs(flags.VendorDir)
	if err != nil {
		return maskAny(err)
	}
	targetDir := flags.TargetDir
	if targetDir == "" {
		targetDir = filepath.Join(gopath, "src")
	}
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return maskAny(err)
	}
	if targetDir != vendorDir {
		if err := copyFromVendor(log, targetDir, vendorDir, "Copying"); err != nil {
			return maskAny(err)
		}
	}
	if err := flattenGoDir(log, targetDir, targetDir, flags.NoRemove); err != nil {
		return maskAny(err)
	}

	return nil
}

func copyFromVendor(log *log.Logger, goDir, vendorDir, verb string) error {
	entries, err := ioutil.ReadDir(vendorDir)
	if err != nil {
		return maskAny(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		entryVendorDir := filepath.Join(vendorDir, entry.Name())
		entryGoDir := filepath.Join(goDir, entry.Name())
		if _, err := os.Stat(entryGoDir); os.IsNotExist(err) {
			// We must create a link
			log.Debugf("%s %s", verb, makeRel(entryVendorDir))
			if err := os.MkdirAll(goDir, 0777); err != nil {
				return maskAny(err)
			}
			if err := util.ExecPrintError(nil, "rsync", "-a", "--ignore-existing", entryVendorDir, goDir); err != nil {
				return maskAny(err)
			}

		} else if err != nil {
			return maskAny(err)
		} else {
			// entry already exists in godir, recurse into the directory
			if err := copyFromVendor(log, entryGoDir, entryVendorDir, verb); err != nil {
				return maskAny(err)
			}
		}
	}

	return nil
}

func flattenGoDir(log *log.Logger, goSrcDir, curDir string, noRemove bool) error {
	vendorDir, err := GetVendorDir(curDir)
	if err != nil {
		return maskAny(err)
	}
	if _, err := os.Stat(vendorDir); err == nil {
		if err := copyFromVendor(log, goSrcDir, vendorDir, "Flattening"); err != nil {
			return maskAny(err)
		}
		if !noRemove {
			if err := os.RemoveAll(vendorDir); err != nil {
				return maskAny(err)
			}
		}
	}

	// Recurse into sub-directories
	entries, err := ioutil.ReadDir(curDir)
	if err != nil {
		return maskAny(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := flattenGoDir(log, goSrcDir, filepath.Join(curDir, entry.Name()), noRemove); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

func GetVendorDir(dir string) (string, error) {
	settings, err := settings.Read(dir)
	if err != nil {
		return "", maskAny(err)
	}
	if settings == nil || settings.GoVendorDir == "" {
		return filepath.Join(dir, DefaultVendorDir), nil
	}
	return filepath.Join(dir, settings.GoVendorDir), nil
}

func makeRel(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil {
		return path
	}
	return rel
}
