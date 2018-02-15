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

	"github.com/pulcy/pulsar/cache"
)

var (
	clearCmd = &cobra.Command{
		Use: "clear",
		Run: UsageFunc,
	}
	clearCacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "Clear a cache folder",
		Long:  "Clear a cache folder",
		Run:   runClearCache,
	}
)

func init() {
	clearCmd.AddCommand(clearCacheCmd)
	mainCmd.AddCommand(clearCmd)
}

func runClearCache(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		if err := cache.ClearAll(); err != nil {
			Quitf("Clear cache failed: %v\n", err)
		}
	} else {
		for _, key := range args {
			if err := cache.Clear(key); err != nil {
				Quitf("Clear cache failed: %v\n", err)
			}
		}
	}
}
