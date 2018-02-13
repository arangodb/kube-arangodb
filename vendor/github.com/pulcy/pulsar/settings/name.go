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

package settings

import (
	"path"
	"path/filepath"

	logging "github.com/op/go-logging"

	"github.com/pulcy/pulsar/git"
	"github.com/pulcy/pulsar/util"
)

// Try to read project name
func GetProjectName(log *logging.Logger, projectDir string) (string, error) {
	var name string
	if err := util.ExecuteInDir(projectDir, func() error {
		if url, err := git.GetRemoteOriginUrl(log); err == nil {
			if info, err := util.ParseVCSURL(url); err != nil {
				return maskAny(err)
			} else {
				name = info.Name
			}
		}
		return nil
	}); err != nil {
		return "", maskAny(err)
	}
	if name != "" {
		return name, nil
	}

	var err error
	projectDir, err = filepath.Abs(projectDir)
	if err != nil {
		return "", maskAny(err)
	}
	return path.Base(projectDir), nil
}
