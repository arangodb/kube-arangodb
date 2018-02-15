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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/juju/errgo"
	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/docker"
	"github.com/pulcy/pulsar/git"
	"github.com/pulcy/pulsar/settings"
	"github.com/pulcy/pulsar/util"
)

const (
	packageJsonFile   = "package.json"
	nameKey           = "name"
	versionKey        = "version"
	makefileFile      = "Makefile"
	gruntfileFile     = "Gruntfile.js"
	dockerfileFile    = "Dockerfile"
	defaultPerm       = 0664
	nodeModulesFolder = "node_modules"
)

type Flags struct {
	ReleaseType    string
	DockerRegistry string
}

func Release(log *log.Logger, flags *Flags) error {
	// Detect environment
	hasMakefile := false
	isDev := flags.ReleaseType == "dev"
	if _, err := os.Stat(makefileFile); err == nil {
		hasMakefile = true
		log.Info("Found %s", makefileFile)
	}

	hasGruntfile := false
	if _, err := os.Stat(gruntfileFile); err == nil {
		hasGruntfile = true
		log.Info("Found %s", gruntfileFile)
	}

	hasDockerfile := false
	if _, err := os.Stat(dockerfileFile); err == nil {
		hasDockerfile = true
		log.Info("Found %s", dockerfileFile)
	}

	// Read the current version and name
	info, err := GetProjectInfo()
	if err != nil {
		return maskAny(err)
	}

	log.Info("Found old version %s", info.Version)
	version, err := semver.NewVersion(info.Version)
	if err != nil {
		return maskAny(err)
	}

	// Check repository state
	if !isDev {
		if err := checkRepoClean(log, info.GitBranch); err != nil {
			return maskAny(err)
		}
	}

	// Bump version
	switch flags.ReleaseType {
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
	case "minor":
		version.Minor++
		version.Patch = 0
	case "patch":
		version.Patch++
	case "dev":
		// Do not change version
	default:
		return errgo.Newf("Unknown release type %s", flags.ReleaseType)
	}
	version.Metadata = ""

	// Write new release version
	if !isDev {
		if err := writeVersion(log, version.String(), info.Manifests, info.GradleConfigFile, false); err != nil {
			return maskAny(err)
		}
	}

	// Build project
	if hasGruntfile && !info.NoGrunt {
		if _, err := os.Stat(nodeModulesFolder); os.IsNotExist(err) {
			log.Info("Folder %s not found", nodeModulesFolder)
			if err := util.ExecPrintError(log, "npm", "install"); err != nil {
				return maskAny(err)
			}
		}
		if err := util.ExecPrintError(log, "grunt", "build-release"); err != nil {
			return maskAny(err)
		}
	}
	if hasMakefile {
		// Clean first
		if !isDev {
			if err := util.ExecPrintError(log, "make", info.Targets.CleanTarget); err != nil {
				return maskAny(err)
			}
		}
		// Now build
		makeArgs := []string{}
		if info.Targets.ReleaseTarget != "" {
			makeArgs = append(makeArgs, info.Targets.ReleaseTarget)
		}
		if err := util.ExecPrintError(log, "make", makeArgs...); err != nil {
			return maskAny(err)
		}
	}

	if hasDockerfile {
		// Build docker images
		tagVersion := version.String()
		if isDev {
			tagVersion = strings.Replace(time.Now().Format("2006-01-02-15-04-05"), "-", "", -1)
		}
		imageAndVersion := fmt.Sprintf("%s:%s", info.Image, tagVersion)
		imageAndMajorVersion := fmt.Sprintf("%s:%d", info.Image, version.Major)
		imageAndMinorVersion := fmt.Sprintf("%s:%d.%d", info.Image, version.Major, version.Minor)
		imageAndLatest := fmt.Sprintf("%s:latest", info.Image)
		buildTag := path.Join(info.Namespace, imageAndVersion)
		buildLatestTag := path.Join(info.Namespace, imageAndLatest)
		buildMajorVersionTag := path.Join(info.Namespace, imageAndMajorVersion)
		buildMinorVersionTag := path.Join(info.Namespace, imageAndMinorVersion)
		if err := util.ExecPrintError(log, "docker", "build", "--tag", buildTag, "."); err != nil {
			return maskAny(err)
		}
		if info.TagLatest {
			util.ExecSilent(log, "docker", "rmi", buildLatestTag)
			if err := util.ExecPrintError(log, "docker", "tag", buildTag, buildLatestTag); err != nil {
				return maskAny(err)
			}
		}
		if info.TagMajorVersion && !isDev {
			util.ExecSilent(log, "docker", "rmi", buildMajorVersionTag)
			if err := util.ExecPrintError(log, "docker", "tag", buildTag, buildMajorVersionTag); err != nil {
				return maskAny(err)
			}
		}
		if info.TagMinorVersion && !isDev {
			util.ExecSilent(log, "docker", "rmi", buildMinorVersionTag)
			if err := util.ExecPrintError(log, "docker", "tag", buildTag, buildMinorVersionTag); err != nil {
				return maskAny(err)
			}
		}
		registry := flags.DockerRegistry
		if info.Registry != "" {
			registry = info.Registry
		}
		namespace := info.Namespace
		if registry != "" || namespace != "" {
			// Push image to registry
			if err := docker.Push(log, imageAndVersion, registry, namespace); err != nil {
				return maskAny(err)
			}
			if info.TagLatest {
				// Push latest image to registry
				if err := docker.Push(log, imageAndLatest, registry, namespace); err != nil {
					return maskAny(err)
				}
			}
			if info.TagMajorVersion && !isDev {
				// Push major version image to registry
				if err := docker.Push(log, imageAndMajorVersion, registry, namespace); err != nil {
					return maskAny(err)
				}
			}
			if info.TagMinorVersion && !isDev {
				// Push minor version image to registry
				if err := docker.Push(log, imageAndMinorVersion, registry, namespace); err != nil {
					return maskAny(err)
				}
			}
		}
	}

	// Build succeeded, re-write new release version and commit
	if !isDev {
		if err := writeVersion(log, version.String(), info.Manifests, info.GradleConfigFile, true); err != nil {
			return maskAny(err)
		}

		// Tag version
		if err := git.Tag(log, version.String()); err != nil {
			return maskAny(err)
		}

		// Create github release (if needed)
		if err := createGithubRelease(log, version.String(), *info); err != nil {
			return maskAny(err)
		}

		// Update version to "+git" working version
		version.Metadata = "git"

		// Write new release version
		if err := writeVersion(log, version.String(), info.Manifests, info.GradleConfigFile, true); err != nil {
			return maskAny(err)
		}

		// Push changes
		if err := git.Push(log, "origin", false); err != nil {
			return maskAny(err)
		}

		// Push tags
		if err := git.Push(log, "origin", true); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

// Update the version of the given package (if any) and an existing VERSION file (if any)
// Commit changes afterwards
func writeVersion(log *log.Logger, version string, manifests []Manifest, gradleConfigFile string, commit bool) error {
	files := []string{}
	for _, mf := range manifests {
		mf.Data[versionKey] = version
		data, err := json.MarshalIndent(mf.Data, "", "  ")
		if err != nil {
			return maskAny(err)
		}
		if err := ioutil.WriteFile(mf.Path, data, defaultPerm); err != nil {
			return maskAny(err)
		}
		files = append(files, mf.Path)
	}
	if _, err := os.Stat(settings.VersionFile); err == nil {
		if err := ioutil.WriteFile(settings.VersionFile, []byte(version), defaultPerm); err != nil {
			return maskAny(err)
		}
		files = append(files, settings.VersionFile)
	}
	if gradleConfigFile != "" {
		if err := createGradleVersionFile(gradleConfigFile, version); err != nil {
			return maskAny(err)
		}
		files = append(files, gradleConfigFile)
	}

	if commit {
		if err := git.Add(log, files...); err != nil {
			return maskAny(err)
		}
		msg := fmt.Sprintf("Updated version to %s", version)
		if err := git.Commit(log, msg); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

// Are the no uncommited changes in this repo?
func checkRepoClean(log *log.Logger, branch string) error {
	if st, err := git.Status(log, true); err != nil {
		return maskAny(err)
	} else if st != "" {
		return maskAny(errgo.New("There are uncommited changes"))
	}
	if err := git.Fetch(log, "origin"); err != nil {
		return maskAny(err)
	}
	if diff, err := git.Diff(log, branch, path.Join("origin", branch)); err != nil {
		return maskAny(err)
	} else if diff != "" {
		return maskAny(errgo.Newf("%s is not in sync with origin", branch))
	}

	return nil
}
