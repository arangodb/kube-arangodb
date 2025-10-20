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

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func licenseInventory() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "inventory [flags] output"
	cmd.Short = "Inventory Generator"

	if err := cli.RegisterFlags(&cmd, flagDeployment); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(licenseInventoryRun).Run

	return &cmd, nil
}

func licenseInventoryRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.Errorf("Invalid arguments")
	}

	inv, err := buildInventory(cmd)
	if err != nil {
		return err
	}

	d, err := ugrpc.Marshal(inv)
	if err != nil {
		return err
	}

	return os.WriteFile(args[0], d, 0600)
}
