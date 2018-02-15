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
	"fmt"
	"os"
	"path/filepath"

	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/git"
	vcsurl "github.com/sourcegraph/go-vcsurl"
)

type GoPathFlags struct {
	Package string // If set, use this package instead of the origin URL from the local repo
}

// CreateLocalGoPath creates a local .gobuild folder with a GOPATH folder structure in it.
func CreateLocalGoPath(log *log.Logger, flags *GoPathFlags) error {
	// Parse repo info
	if flags.Package == "" {
		remote, err := git.GetRemoteOriginUrl(log)
		if err != nil {
			return maskAny(err)
		}
		flags.Package = remote
	}
	gitURL, err := vcsurl.Parse(flags.Package)
	if err != nil {
		return maskAny(err)
	}
	// Prepare dirs
	curDir, err := os.Getwd()
	if err != nil {
		return maskAny(err)
	}
	gobuildDir := filepath.Join(curDir, ".gobuild")
	orgDir := filepath.Join(gobuildDir, "src", string(gitURL.RepoHost), gitURL.Username)
	repoDir := filepath.Join(orgDir, gitURL.Name)
	relRepoDir, err := filepath.Rel(orgDir, curDir)
	targetDir := curDir
	if err == nil {
		targetDir = relRepoDir
	}
	if _, err := os.Stat(repoDir); err != nil {
		if err := os.MkdirAll(orgDir, 0755); err != nil {
			return maskAny(err)
		}
		if err := os.Symlink(targetDir, repoDir); err != nil {
			return maskAny(err)
		}
	}

	fmt.Println(gobuildDir)

	return nil
}
