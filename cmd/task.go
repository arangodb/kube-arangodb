//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	var deploymentName string

	cmdMain.AddCommand(cmdTask)
	cmdOps.AddCommand(cmdTask)

	cmdTask.AddCommand(cmdTaskCreate)
	cmdTaskCreate.Flags().StringVarP(&deploymentName, ArgDeploymentName, "d", "",
		"Name of ArangoDeployment for which Task will be created - necessary when more than one deployment exist within one namespace")

	cmdTask.AddCommand(cmdTaskState)
}

var cmdTask = &cobra.Command{
	Use: "task",
	Run: executeUsage,
}

var cmdTaskCreate = &cobra.Command{
	Use:   "create",
	Short: "Create task",
	Run:   taskCreate,
}

var cmdTaskState = &cobra.Command{
	Use:   "state",
	Short: "Get Task state",
	Long:  "It prints the task current state on the stdout",
	Run:   taskState,
}

func taskCreate(cmd *cobra.Command, args []string) {
	logger.Info("TODO: create task")
}

func taskState(cmd *cobra.Command, args []string) {
	logger.Info("TODO: check task state")
}
