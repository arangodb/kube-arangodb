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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/juju/errgo"
	"github.com/mgutz/ansi"
	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/cache"
	"github.com/pulcy/pulsar/util"
)

const (
	srcDir     = "src"
	cacheValid = time.Hour * 12
)

var (
	maskAny   = errgo.MaskFunc(errgo.Any)
	allGood   = ansi.ColorFunc("")
	updating  = ansi.ColorFunc("cyan")
	attention = ansi.ColorFunc("yellow")
	envMutex  sync.Mutex
	gopath    string
)

type GetFlags struct {
	Package string
}

func init() {
	gopath = os.Getenv("GOPATH")
}

// Get executes a `go get` with a cache support.
func Get(log *log.Logger, flags *GetFlags) error {
	// Check GOPATH
	if gopath == "" {
		return maskAny(errors.New("Specify GOPATH"))
	}
	gopathDir := strings.Split(gopath, string(os.PathListSeparator))[0]

	// Get cache dir
	cachedir, cacheIsValid, err := cache.Dir(flags.Package, cacheValid)
	if err != nil {
		return maskAny(err)
	}

	if !cacheIsValid {
		// Cache has become invalid
		log.Info(updating("Refreshing cache of %s"), flags.Package)
		// Execute `go get` towards the cache directory
		if err := runGoGet(log, flags.Package, cachedir); err != nil {
			return maskAny(err)
		}
	}

	// Sync with local gopath
	if err := os.MkdirAll(gopathDir, 0777); err != nil {
		return maskAny(err)
	}
	if err := util.ExecPrintError(nil, "rsync", "-a", filepath.Join(cachedir, srcDir), gopathDir); err != nil {
		return maskAny(err)
	}

	return nil
}

func runGoGet(log *log.Logger, pkg, gopath string) error {
	envMutex.Lock()
	defer envMutex.Unlock()

	return func() error {
		// Restore GOPATH on exit
		defer os.Setenv("GOPATH", gopath)
		// Set GOPATH
		if err := os.Setenv("GOPATH", gopath); err != nil {
			return maskAny(err)
		}
		//log.Info("GOPATH=%s", gopath)
		return maskAny(util.ExecPrintError(log, "go", "get", pkg))
	}()

}
