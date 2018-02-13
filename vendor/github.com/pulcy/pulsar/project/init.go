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

package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/op/go-logging"
	"github.com/pulcy/pulsar/settings"
	"github.com/pulcy/pulsar/util"
)

const (
	ProjectTypeGo = "go"

	gitDirPath    = ".git"
	gitIgnorePath = ".gitignore"
	envrcPath     = ".envrc"

	initialGitIgnore = `.gobuild
.DS_Store`
	initialEnvrc = `export GOPATH=$(pwd)/.gobuild
PATH_add $GOPATH/bind`
)

type InitializeFlags struct {
	ProjectDir  string
	ProjectType string
}

func Initialize(log *log.Logger, flags InitializeFlags) error {
	if flags.ProjectDir == "" {
		flags.ProjectDir = "."
	}
	var err error
	flags.ProjectDir, err = filepath.Abs(flags.ProjectDir)
	if err != nil {
		return maskAny(err)
	}
	// Ensure directory exists
	os.MkdirAll(flags.ProjectDir, 0755)
	// Ensure git is initialized
	if err := initGit(log, flags.ProjectDir); err != nil {
		return maskAny(err)
	}
	// Ensure .gitignore is initialized
	if err := initGitIgnore(log, flags.ProjectDir); err != nil {
		return maskAny(err)
	}
	// Ensure VERSION exists
	if err := initVersion(log, flags.ProjectDir); err != nil {
		return maskAny(err)
	}
	// Ensure Makefile exists
	if err := initMakefile(log, flags.ProjectDir, flags.ProjectType); err != nil {
		return maskAny(err)
	}
	// Ensure .envrc exists
	if err := initEnvrc(log, flags.ProjectDir, flags.ProjectType); err != nil {
		return maskAny(err)
	}
	return nil
}

func initGit(log *log.Logger, projectDir string) error {
	path := filepath.Join(projectDir, gitDirPath)
	if info, err := os.Stat(path); os.IsNotExist(err) {
		if err := util.ExecuteInDir(projectDir, func() error {
			output, err := util.Exec(log, "git", "init")
			if err != nil {
				log.Error(output)
				return maskAny(err)
			}
			return nil
		}); err != nil {
			return maskAny(err)
		}
	} else if err != nil {
		return maskAny(err)
	} else if !info.IsDir() {
		return maskAny(fmt.Errorf("%s must be a directory", path))
	} else {
		log.Debugf("Git already initialized in %s", projectDir)
	}
	return nil
}

func initVersion(log *log.Logger, projectDir string) error {
	path := filepath.Join(projectDir, settings.VersionFile)
	if info, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating %s", settings.VersionFile)
		if err := ioutil.WriteFile(path, []byte(settings.InitialVersion), 0644); err != nil {
			return maskAny(err)
		}
		return nil
	} else if err != nil {
		return maskAny(err)
	} else if info.IsDir() {
		return maskAny(fmt.Errorf("%s must be a file", path))
	} else {
		log.Debugf("%s already initialized in %s", settings.VersionFile, projectDir)
		return nil
	}
}

func initGitIgnore(log *log.Logger, projectDir string) error {
	path := filepath.Join(projectDir, gitIgnorePath)
	if info, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating %s", gitIgnorePath)
		if err := ioutil.WriteFile(path, []byte(initialGitIgnore), 0644); err != nil {
			return maskAny(err)
		}
		return nil
	} else if err != nil {
		return maskAny(err)
	} else if info.IsDir() {
		return maskAny(fmt.Errorf("%s must be a file", path))
	} else {
		log.Debugf("%s already initialized in %s", gitIgnorePath, projectDir)
		return nil
	}
}

func initEnvrc(log *log.Logger, projectDir, projectType string) error {
	if projectType != ProjectTypeGo {
		return nil
	}
	path := filepath.Join(projectDir, envrcPath)
	if info, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating %s", gitIgnorePath)
		if err := ioutil.WriteFile(path, []byte(initialEnvrc), 0644); err != nil {
			return maskAny(err)
		}
		return nil
	} else if err != nil {
		return maskAny(err)
	} else if info.IsDir() {
		return maskAny(fmt.Errorf("%s must be a file", path))
	} else {
		log.Debugf("%s already initialized in %s", envrcPath, projectDir)
		return nil
	}
}
