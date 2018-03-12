//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coreos/go-semver/semver"
)

var (
	versionFile string // Full path of VERSION file
	releaseType string // What type of release to create (major|minor|patch)
	ghRelease   string // Full path of github-release tool
	ghUser      string // Github account name to create release in
	ghRepo      string // Github repository name to create release in
)

func init() {
	flag.StringVar(&versionFile, "versionfile", "./VERSION", "Path of the VERSION file")
	flag.StringVar(&releaseType, "type", "patch", "Type of release to build (major|minor|patch)")
	flag.StringVar(&ghRelease, "github-release", ".gobuild/bin/github-release", "Full path of github-release tool")
	flag.StringVar(&ghUser, "github-user", "arangodbdb", "Github account name to create release in")
	flag.StringVar(&ghRepo, "github-repo", "k8s-operator", "Github repository name to create release in")
}

func main() {
	flag.Parse()
	ensureGithubToken()
	checkCleanRepo()
	version := bumpVersion(releaseType)
	make("clean", nil)
	make("all", map[string]string{
		"DOCKERNAMESPACE": "arangodb",
		"IMAGETAG":        version,
		"MANIFESTPATH":    "manifests/arango-operator.yaml",
	})
	gitCommitAll(fmt.Sprintf("Updated manifest to %s", version)) // Commit manifest
	gitTag(version)
	githubCreateRelease(version)
	bumpVersion("devel")
}

// ensureGithubToken makes sure the GITHUB_TOKEN envvar is set.
func ensureGithubToken() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		p := filepath.Join(os.Getenv("HOME"), ".arangodb/github-token")
		if raw, err := ioutil.ReadFile(p); err != nil {
			log.Fatalf("Failed to release '%s': %v", p, err)
		} else {
			token = strings.TrimSpace(string(raw))
			os.Setenv("GITHUB_TOKEN", token)
		}
	}
}

func checkCleanRepo() {
	output, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		log.Fatalf("Failed to check git status: %v\n", err)
	}
	if strings.TrimSpace(string(output)) != "" {
		log.Fatal("Repository has uncommitted changes\n")
	}
}

func make(target string, envVars map[string]string) {
	if err := run("make", []string{target}, envVars); err != nil {
		log.Fatalf("Failed to make %s: %v\n", target, err)
	}
}

func bumpVersion(action string) string {
	contents, err := ioutil.ReadFile(versionFile)
	if err != nil {
		log.Fatalf("Cannot read '%s': %v\n", versionFile, err)
	}
	version := semver.New(strings.TrimSpace(string(contents)))

	switch action {
	case "patch":
		version.BumpPatch()
	case "minor":
		version.BumpMinor()
	case "major":
		version.BumpMajor()
	case "devel":
		version.Metadata = "git"
	}
	contents = []byte(version.String())

	if err := ioutil.WriteFile(versionFile, contents, 0755); err != nil {
		log.Fatalf("Cannot write '%s': %v\n", versionFile, err)
	}

	gitCommitAll(fmt.Sprintf("Updated to %s", version))
	log.Printf("Updated '%s' to '%s'\n", versionFile, string(contents))

	return version.String()
}

func gitCommitAll(message string) {
	args := []string{
		"commit",
		"--all",
		"-m", message,
	}
	if err := run("git", args, nil); err != nil {
		log.Fatalf("Failed to commit: %v\n", err)
	}
	if err := run("git", []string{"push"}, nil); err != nil {
		log.Fatalf("Failed to push commit: %v\n", err)
	}
}

func gitTag(version string) {
	if err := run("git", []string{"tag", version}, nil); err != nil {
		log.Fatalf("Failed to tag: %v\n", err)
	}
	if err := run("git", []string{"push", "--tags"}, nil); err != nil {
		log.Fatalf("Failed to push tags: %v\n", err)
	}
}

func githubCreateRelease(version string) {
	// Create draft release
	args := []string{
		"release",
		"--user", ghUser,
		"--repo", ghRepo,
		"--tag", version,
		"--draft",
	}
	if err := run(ghRelease, args, nil); err != nil {
		log.Fatalf("Failed to create github release: %v\n", err)
	}
	// Finalize release
	args = []string{
		"edit",
		"--user", ghUser,
		"--repo", ghRepo,
		"--tag", version,
	}
	if err := run(ghRelease, args, nil); err != nil {
		log.Fatalf("Failed to finalize github release: %v\n", err)
	}
}

func run(cmd string, args []string, envVars map[string]string) error {
	c := exec.Command(cmd, args...)
	if envVars != nil {
		for k, v := range envVars {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
