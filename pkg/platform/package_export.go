//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package platform

import (
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/platform/pack"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func packageExport() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "export [flags] package output"
	cmd.Short = "Export the package in the ZIP Format"

	if err := cli.RegisterFlags(&cmd, flagLicenseManager, flagRegistry); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageExportRun).Run

	return &cmd, nil
}

func packageExportRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.Errorf("Invalid arguments")
	}

	pkg, err := getHelmPackages(args[0])
	if err != nil {
		logger.Err(err).Error("Unable to read the file")
		return err
	}

	out := args[1]

	rc, err := flagRegistry.Client(cmd, flagLicenseManager)
	if err != nil {
		return err
	}

	endpoint, err := flagLicenseManager.Endpoint(cmd)
	if err != nil {
		return err
	}

	return pack.Export(cmd.Context(), pack.NewCache("cache"), endpoint, out, rc, pkg)
}
