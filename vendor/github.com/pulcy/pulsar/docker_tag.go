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
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"

	"github.com/pulcy/pulsar/release"
)

var (
	dockerTagImage string
	dockerTagCmd   = &cobra.Command{
		Use:   "docker-tag",
		Short: "Get the docker tag for the current project",
		Long:  "Returns the image:tag for the current project",
		Run:   runDockerTag,
	}
)

func init() {
	dockerTagCmd.Flags().StringVarP(&dockerTagImage, "image", "i", "", "Docker image name")
	mainCmd.AddCommand(dockerTagCmd)
}

func runDockerTag(cmd *cobra.Command, args []string) {
	info, err := release.GetProjectInfo()
	if err != nil {
		Quitf("%s\n", err)
	}
	version, err := semver.NewVersion(info.Version)
	if err != nil {
		Quitf("%s\n", err)
	}
	tag := version.String()
	if version.Metadata != "" {
		tag = "latest"
	}
	if dockerTagImage == "" {
		dockerTagImage = info.Image
	}
	Printf("%s:%s", dockerTagImage, tag)
}
