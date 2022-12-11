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
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/debug_package"
)

func init() {
	cmdMain.AddCommand(debugPackage)

	f := debugPackage.Flags()

	f.StringVarP(&debugPackageInput.Output, "output", "o", "out.tar.gz", "Output of the result gz file. If set to `-` then stdout is used")

	debug_package.InitCommand(debugPackage)
}

var debugPackage = &cobra.Command{
	Use:   "debugPackage",
	Short: "[WiP] Generate debug package for debugging",
	RunE:  debugPackageFunc,
}

var debugPackageInput struct {
	Output string
}

func debugPackageFunc(cmd *cobra.Command, _ []string) error {
	if debugPackageInput.Output == "-" {
		return debugPackageStdOut(cmd)
	}

	return debugPackageFile(cmd)
}

func debugPackageStdOut(cmd *cobra.Command) (returnError error) {
	return debugPackageGZip(cmd, os.Stdout)
}

func debugPackageFile(cmd *cobra.Command) (returnError error) {
	out, err := os.OpenFile("./out.tar.gz", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := out.Close(); err != nil {
			if returnError == nil {
				returnError = err
			}
		}
	}()

	return debugPackageGZip(cmd, out)
}

func debugPackageGZip(cmd *cobra.Command, out io.Writer) (returnError error) {
	gw := gzip.NewWriter(out)

	defer func() {
		if err := gw.Close(); err != nil {
			if returnError == nil {
				returnError = err
			}
		}
	}()

	return debugPackageRaw(cmd, gw)
}

func debugPackageRaw(cmd *cobra.Command, gw io.Writer) error {
	return debug_package.GenerateD(cmd, gw)
}
