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
	"strings"

	logging "github.com/op/go-logging"

	"github.com/pulcy/pulsar/git"
	"github.com/pulcy/pulsar/util"
)

// Try to return the URL of a project's sibling project (e.g. 'git@github.com:pulcy/sibling.git')
func GetProjectSiblingURL(log *logging.Logger, projectDir, siblingName string) (string, error) {
	var siblingURL string
	if err := util.ExecuteInDir(projectDir, func() error {
		if url, err := git.GetRemoteOriginUrl(log); err == nil {
			if info, err := util.ParseVCSURL(url); err != nil {
				return maskAny(err)
			} else {
				siblingFullName := path.Join(path.Dir(info.FullName), siblingName)
				siblingURL = strings.Replace(url, info.FullName, siblingFullName, 1)
				// fmt.Printf("url=%s\n", url)
				// fmt.Printf("siblingFullName=%s\n", siblingFullName)
				// fmt.Printf("info.FullName=%s\n", info.FullName)
			}
		}
		return nil
	}); err != nil {
		return "", maskAny(err)
	}
	return siblingURL, nil
}
