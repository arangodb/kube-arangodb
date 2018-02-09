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

package release

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/op/go-logging"
	"github.com/pulcy/pulsar/git"
	vcsurl "gopkg.in/sourcegraph/go-vcsurl.v1"

	"github.com/pulcy/pulsar/github"
)

func createGithubRelease(log *log.Logger, version string, info ProjectInfo) error {
	// Are assets specified?
	if len(info.GithubAssets) == 0 {
		log.Debugf("No github-assets specified, no github release is created")
		return nil
	}

	// Check existance of all assets
	for _, asset := range info.GithubAssets {
		if _, err := os.Stat(asset.RelPath); err != nil {
			return maskAny(fmt.Errorf("Cannot stat asset '%s': %v", asset.RelPath, err))
		}
	}

	// Is the repository URL suitable for github releases?
	url, err := git.GetRemoteOriginUrl(log)
	if err != nil {
		return maskAny(err)
	}
	repoInfo, err := vcsurl.Parse(url)
	if err != nil {
		return maskAny(err)
	}
	if repoInfo.RepoHost != vcsurl.GitHub || repoInfo.VCS != vcsurl.Git {
		return maskAny(fmt.Errorf("Cannot create github-release because repository is not a git repo or not hosted on github"))
	}

	// Load github token
	token, err := github.LoadGithubToken()
	if err != nil {
		return maskAny(err)
	}
	gs := github.GithubService{
		Logger:     log,
		Token:      token,
		User:       repoInfo.Username,
		Repository: repoInfo.Name,
	}

	// Create github release
	relOpt := github.ReleaseCreate{
		TagName: version,
		Name:    fmt.Sprintf("v%s", version),
	}
	if err := gs.CreateRelease(relOpt); err != nil {
		return maskAny(err)
	}

	// Attach assets
	for _, asset := range info.GithubAssets {
		opt := github.UploadAssetOptions{
			TagName:  version,
			FileName: filepath.Base(asset.RelPath),
			Label:    asset.Label,
			Path:     asset.RelPath,
		}
		if err := gs.UploadAsset(opt); err != nil {
			return maskAny(err)
		}
	}

	// Update tags
	if err := git.FetchTags(log, "origin"); err != nil {
		return maskAny(err)
	}

	return nil
}
