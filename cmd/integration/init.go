//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package integration

import (
	goflag "flag"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	integrations "github.com/arangodb/kube-arangodb/pkg/integrations"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

var (
	cmd = cobra.Command{
		Use:  "arangodb_operator_integration",
		RunE: cli.Usage,
	}
)

func init() {
	if err := integrations.Register(&cmd); err != nil {
		panic(err.Error())
	}
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func Command() *cobra.Command {
	return &cmd
}

func Execute() int {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := cmd.ExecuteContext(shutdown.Context()); err != nil {
		return 1
	}

	return 0
}
