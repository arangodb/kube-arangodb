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
	"os"

	logPkg "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var (
	projectVersion = "dev"
	projectName    = "pulsar"
	projectBuild   = "dev"
	log            = logPkg.MustGetLogger(projectName)

	mainCmd = &cobra.Command{
		Use:   projectName,
		Short: "Pulsar is a helper for Pulcy development environments",
		Run:   UsageFunc,
	}
)

func main() {
	mainCmd.Execute()
}

func Printf(message string, args ...interface{}) {
	fmt.Printf(message, args...)
}

func Quitf(message string, args ...interface{}) {
	Printf(message, args...)
	os.Exit(1)
}

// Print if quiet flag has not been set
func Infof(message string, args ...interface{}) {
	fmt.Printf(message, args...)
}

func CommandError(c *cobra.Command, prefix string, args ...interface{}) {
	prefix = fmt.Sprintf(prefix, args...)
	Quitf("%sUsage: %s\n", prefix, c.CommandPath())
}

func UsageFunc(cmd *cobra.Command, args []string) {
	cmd.Help()
}
