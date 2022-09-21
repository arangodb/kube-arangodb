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
	"compress/gzip"
	"os"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/debug_package"
)

func init() {
	cmdMain.AddCommand(debugPackage)

	debug_package.InitCommand(debugPackage)
}

var debugPackage = &cobra.Command{
	Use:   "debugPackage",
	Short: "[WiP] Generate debug package for debugging",
	RunE:  debugPackageFunc,
}

func debugPackageFunc(cmd *cobra.Command, _ []string) error {
	f, err := os.OpenFile("./out.tar.gz", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	gw := gzip.NewWriter(f)

	if err := debug_package.GenerateD(cmd, gw); err != nil {
		return err
	}

	if err := gw.Close(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
