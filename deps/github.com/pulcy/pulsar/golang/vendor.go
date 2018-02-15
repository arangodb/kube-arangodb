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
	"os"
	"path/filepath"
	"time"

	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/cache"
	"github.com/pulcy/pulsar/util"
)

const (
	DefaultVendorDir = "vendor"
)

type VendorFlags struct {
	Package   string
	VendorDir string
}

// Get executes a `go get` with a cache support.
func Vendor(log *log.Logger, flags *VendorFlags) error {
	// Get cache dir
	cachedir, _, err := cache.Dir(flags.Package, time.Millisecond)
	if err != nil {
		return maskAny(err)
	}

	// Cache has become invalid
	log.Info(updating("Fetching %s"), flags.Package)
	// Execute `go get` towards the cache directory
	if err := runGoGet(log, flags.Package, cachedir); err != nil {
		return maskAny(err)
	}

	// Sync with vendor dir
	if err := os.MkdirAll(flags.VendorDir, 0777); err != nil {
		return maskAny(err)
	}
	if err := util.ExecPrintError(nil, "rsync", "--exclude", ".git", "-a", filepath.Join(cachedir, srcDir)+"/", flags.VendorDir); err != nil {
		return maskAny(err)
	}

	return nil
}
