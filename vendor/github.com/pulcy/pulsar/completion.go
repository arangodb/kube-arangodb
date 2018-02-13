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
)

var (
	commandListCmd = &cobra.Command{
		Use:   "command-list",
		Short: "Gets all commands",
		Run:   runCommandListCmd,
	}
	completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generate a bash completion script",
		Run:   runCompletionCmd,
	}
)

func init() {
	mainCmd.AddCommand(commandListCmd)
	mainCmd.AddCommand(completionCmd)
}

func runCommandListCmd(cmd *cobra.Command, args []string) {
	for _, c := range mainCmd.Commands() {
		Printf("%s ", c.Name())
	}
	Printf("\n")
}

func runCompletionCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Quitf("Provide a filename argument\n")
	}
	if err := mainCmd.GenBashCompletionFile(args[0]); err != nil {
		Quitf("Failed to generate BASH completion script: %#v\n", err)
	}
}
