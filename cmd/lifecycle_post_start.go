//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/collect"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

var (
	// cmdLifecyclePostStart is the parent for postStart lifecycle hooks (`lifecycle postStart ...`).
	cmdLifecyclePostStart = &cobra.Command{
		Use:    "postStart",
		RunE:   cli.Usage,
		Hidden: true,
	}

	// cmdLifecyclePostStartCollector runs the collector as a postStart hook (`lifecycle postStart collector`).
	cmdLifecyclePostStartCollector = &cobra.Command{
		Use:    "collector",
		Hidden: true,
	}

	// lifecyclePostStartCollectorOptions holds the flags bound to the collector command.
	lifecyclePostStartCollectorOptions collect.Options
)

func init() {
	cmdLifecyclePostStartCollector.RunE = cmdLifecyclePostStartCollectorRunE

	f := cmdLifecyclePostStartCollector.Flags()
	f.DurationVar(&lifecyclePostStartCollectorOptions.Interval, "interval", collect.DefaultInterval, "Collector retry interval")
	f.DurationVar(&lifecyclePostStartCollectorOptions.Timeout, "timeout", collect.DefaultTimeout, "Collector run timeout")

	cmdLifecyclePostStart.AddCommand(cmdLifecyclePostStartCollector)
	cmdLifecycle.AddCommand(cmdLifecyclePostStart)
}

// cmdLifecyclePostStartCollectorRunE delegates to the collector implementation in pkg/collect.
func cmdLifecyclePostStartCollectorRunE(cmd *cobra.Command, _ []string) error {
	return collect.PostStart(cmd.Context(), lifecyclePostStartCollectorOptions)
}
