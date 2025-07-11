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
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/platform/pack"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func packageImport() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "import [flags] registry package output"
	cmd.Short = "Imports the package from the ZIP format"

	if err := cli.RegisterFlags(&cmd, flagRegistryUseCredentials, flagRegistryInsecure, flagRegistryList); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageImportRun).Run

	return &cmd, nil
}

func packageImportRun(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
		return errors.Errorf("Invalid arguments")
	}

	reg := args[0]
	dest := args[1]
	out := args[2]

	rc, err := getRegClient(cmd)
	if err != nil {
		return err
	}

	_, pkg, err := pack.Import(cmd.Context(), dest, rc, reg)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(pkg)
	if err != nil {
		return err
	}

	return os.WriteFile(out, data, 0644)
}
