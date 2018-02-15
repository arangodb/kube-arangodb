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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulcy/pulsar/git"
	"github.com/pulcy/pulsar/project"
	"github.com/pulcy/pulsar/settings"
)

var (
	projectDir  string
	projectType string
	projectCmd  = &cobra.Command{
		Use:   "project",
		Short: "Project helpers",
		Run:   UsageFunc,
	}
	projectCommitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Output project git commit",
		Run:   runProjectCommit,
	}
	projectInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a project folder",
		Run:   runProjectInit,
	}
	projectNameCmd = &cobra.Command{
		Use:   "name",
		Short: "Output project name",
		Run:   runProjectName,
	}
	projectOrganizationCmd = &cobra.Command{
		Use:   "organization",
		Short: "Project organization helpers",
		Run:   UsageFunc,
	}
	projectOrganizationPathCmd = &cobra.Command{
		Use:   "path",
		Short: "Output project organization path (e.g. 'github.com/pulcy')",
		Run:   runProjectOrganizationPath,
	}
	projectOrganizationNameCmd = &cobra.Command{
		Use:   "name",
		Short: "Output project organization name (e.g. 'pulcy')",
		Run:   runProjectOrganizationName,
	}
	projectSiblingCmd = &cobra.Command{
		Use:   "sibling",
		Short: "Project sibling helpers",
		Run:   UsageFunc,
	}
	projectSiblingURLCmd = &cobra.Command{
		Use:   "url",
		Short: "Output clone URL of a sibling of the project (e.g. 'git@github.com:pulcy/sibling.git')",
		Run:   runProjectSiblingURL,
	}
	projectVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Output project version",
		Run:   runProjectVersion,
	}
)

func init() {
	projectCmd.PersistentFlags().StringVarP(&projectDir, "dir", "d", ".", "Project directory")
	projectInitCmd.Flags().StringVar(&projectType, "type", project.ProjectTypeGo, "Project type")

	mainCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCommitCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectNameCmd)
	projectCmd.AddCommand(projectOrganizationCmd)
	projectCmd.AddCommand(projectSiblingCmd)
	projectCmd.AddCommand(projectVersionCmd)

	projectOrganizationCmd.AddCommand(projectOrganizationPathCmd)
	projectOrganizationCmd.AddCommand(projectOrganizationNameCmd)

	projectSiblingCmd.AddCommand(projectSiblingURLCmd)
}

func initProjectDir(args []string) {
	switch len(args) {
	case 0:
	// Do nothing
	case 1:
		if !projectCmd.PersistentFlags().Changed("dir") {
			projectDir = args[0]
		} else {
			Quitf("Cannot set --dir and provide an argument\n")
		}
	default:
		Quitf("Too many arguments\n")
	}
}

func runProjectCommit(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	commit, err := git.GetLatestLocalCommit(nil, projectDir, "", true)
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(commit)
}

func runProjectName(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	name, err := settings.GetProjectName(nil, projectDir)
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(name)
}

func runProjectOrganizationName(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	name, err := settings.GetProjectOrganizationName(nil, projectDir)
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(name)
}

func runProjectOrganizationPath(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	name, err := settings.GetProjectOrganizationPath(nil, projectDir)
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(name)
}

func runProjectSiblingURL(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Quitf("sibling-name required\n")
	}
	url, err := settings.GetProjectSiblingURL(nil, projectDir, args[0])
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(url)
}

func runProjectVersion(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	version, err := settings.ReadVersion(projectDir)
	if err != nil {
		Quitf("%s\n", err)
	}
	fmt.Println(version)
}

func runProjectInit(cmd *cobra.Command, args []string) {
	initProjectDir(args)
	err := project.Initialize(log, project.InitializeFlags{
		ProjectDir:  projectDir,
		ProjectType: projectType,
	})
	if err != nil {
		Quitf("%s\n", err)
	}
}
