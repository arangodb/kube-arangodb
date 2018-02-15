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
	"path"

	"github.com/pulcy/pulsar/settings"
)

type ProjectInfo struct {
	Name             string
	Version          string
	Manifests        []Manifest
	Image            string
	Registry         string
	Namespace        string
	NoGrunt          bool   // If set, grunt won't be called even if there is a Gruntfile.js
	TagLatest        bool   // If set, a latest tag will be set of the docker image
	TagMajorVersion  bool   // If set, a tag will be set to the major version of the docker image (e.g. myimage:3)
	TagMinorVersion  bool   // If set, a tag will be set to the minor version of the docker image (e.g. myimage:3.2)
	GitBranch        string // If set, this branch is expected (defaults to master)
	GradleConfigFile string
	Targets          struct {
		CleanTarget   string
		ReleaseTarget string
	}
	GithubAssets []settings.GithubAsset // If set, creates a github release with given assets.
}

func GetProjectInfo() (*ProjectInfo, error) {
	// Read the current version and name
	project := ""
	manifests := []Manifest{}
	mf, err := tryReadManifest(packageJsonFile)
	if err != nil {
		return nil, maskAny(err)
	}
	var oldVersion string
	if mf != nil {
		manifests = append(manifests, *mf)
		oldVersion = mf.Data[versionKey].(string)
		project = mf.Data[nameKey].(string)
	}
	if oldVersion == "" {
		// Read version from VERSION file
		oldVersion, err = settings.ReadVersion(".")
		if err != nil {
			return nil, maskAny(err)
		}
	}
	if oldVersion == "" {
		oldVersion = "0.0.1"
	}
	if project == "" {
		// Take current directory as name
		if dir, err := os.Getwd(); err != nil {
			return nil, maskAny(err)
		} else {
			project = path.Base(dir)
		}
	}

	// Read project settings (if any)
	image := project
	registry := ""
	namespace := ""
	noGrunt := false
	tagLatest := false
	tagMajorVersion := false
	tagMinorVersion := false
	gradleConfigFile := ""
	gitBranch := "master"
	var githubAssets []settings.GithubAsset
	settings, err := settings.Read(".")
	if err != nil {
		return nil, maskAny(err)
	}
	if settings != nil {
		if settings.Image != "" {
			image = settings.Image
		}
		if settings.Registry != "" {
			registry = settings.Registry
		}
		if settings.Namespace != "" {
			namespace = settings.Namespace
		}
		noGrunt = settings.NoGrunt
		tagLatest = settings.TagLatest
		tagMajorVersion = settings.TagMajorVersion
		tagMinorVersion = settings.TagMinorVersion
		gradleConfigFile = settings.GradleConfigFile
		if settings.GitBranch != "" {
			gitBranch = settings.GitBranch
		}
		githubAssets = settings.GithubAssets

		for _, path := range settings.ManifestFiles {
			mf, err := tryReadManifest(path)
			if err != nil {
				return nil, maskAny(err)
			} else if mf == nil {
				return nil, maskAny(fmt.Errorf("manifest '%s' not found", path))
			}
			manifests = append(manifests, *mf)
		}
	}

	result := &ProjectInfo{
		Name:             project,
		Image:            image,
		Registry:         registry,
		Namespace:        namespace,
		NoGrunt:          noGrunt,
		TagLatest:        tagLatest,
		TagMajorVersion:  tagMajorVersion,
		TagMinorVersion:  tagMinorVersion,
		GitBranch:        gitBranch,
		Version:          oldVersion,
		Manifests:        manifests,
		GradleConfigFile: gradleConfigFile,
		GithubAssets:     githubAssets,
	}
	result.Targets.CleanTarget = "clean"
	if settings != nil && settings.Targets.CleanTarget != "" {
		result.Targets.CleanTarget = settings.Targets.CleanTarget
	}
	result.Targets.ReleaseTarget = ""
	if settings != nil && settings.Targets.ReleaseTarget != "" {
		result.Targets.ReleaseTarget = settings.Targets.ReleaseTarget
	}

	return result, nil
}
