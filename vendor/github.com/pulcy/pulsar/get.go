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

package main

import (
	"github.com/spf13/cobra"

	"github.com/pulcy/pulsar/get"
)

var (
	getFlags = &get.Flags{}
	getCmd   = &cobra.Command{
		Use:   "get",
		Short: "Clone a repo into a folder",
		Long:  "Clone a repo into a folder, checking it out to a specific version",
		Run:   runGet,
	}
)

func init() {
	getCmd.Flags().StringVarP(&getFlags.Version, "version", "b", "", "Specify checkout version")
	getCmd.Flags().BoolVarP(&getFlags.AllowLink, "link", "l", false, "Allow linking to local sibling")
	mainCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) {
	switch len(args) {
	case 2:
		getFlags.RepoUrl = args[0]
		getFlags.Folder = args[1]
		if err := get.Get(log, getFlags); err != nil {
			Quitf("Get failed: %v\n", err)
		}
	default:
		CommandError(cmd, "Expected <repo-url> <folder> arguments\n")
	}
}
