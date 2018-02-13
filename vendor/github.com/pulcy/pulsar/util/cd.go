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

package util

import "os"

// ExecuteInDir changes the current directory to the given folder, executes the given action and then changes
// the current directory back to the original.
func ExecuteInDir(folder string, action func() error) error {
	wd, err := os.Getwd()
	if err != nil {
		return maskAny(err)
	}
	if err := os.Chdir(folder); err != nil {
		return maskAny(err)
	}
	defer os.Chdir(wd)

	if err := action(); err != nil {
		return maskAny(err)
	}
	return nil
}
