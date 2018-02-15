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

	"github.com/pulcy/pulsar/release"
)

var (
	releaseFlags = &release.Flags{}
	releaseCmd   = &cobra.Command{
		Use:   "release",
		Short: "Release a project",
		Long:  "Release a project. Update version, create tags, push etc",
		Run:   UsageFunc,
	}
)

func init() {
	releaseCmd.Flags().StringVarP(&releaseFlags.DockerRegistry, "registry", "r", defaultDockerRegistry(), "Specify docker registry")
	mainCmd.AddCommand(releaseCmd)
	releaseCmd.AddCommand(&cobra.Command{
		Use:   "major",
		Short: "Create a major update",
		Run:   runRelease,
	})
	releaseCmd.AddCommand(&cobra.Command{
		Use:   "minor",
		Short: "Create a minor update",
		Run:   runRelease,
	})
	releaseCmd.AddCommand(&cobra.Command{
		Use:   "patch",
		Short: "Create a patch",
		Run:   runRelease,
	})
	releaseCmd.AddCommand(&cobra.Command{
		Use:   "dev",
		Short: "Create a dev release",
		Run:   runRelease,
	})
}

func runRelease(cmd *cobra.Command, args []string) {
	switch len(args) {
	case 0:
		releaseFlags.ReleaseType = cmd.Name()
		if err := release.Release(log, releaseFlags); err != nil {
			Quitf("Release failed: %v\n", err)
		} else {
			Infof("Release completed\n")
		}
	default:
		CommandError(cmd, "Too many arguments\n")
	}
}
