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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	lmanager "github.com/arangodb/kube-arangodb/pkg/license_manager"
	"github.com/arangodb/kube-arangodb/pkg/platform/inventory"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func licenseGenerate() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "generate"
	cmd.Short = "Generate the License"

	if err := cli.RegisterFlags(&cmd, flagLicenseManager, flagDeploymentID, flagInventory); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(licenseGenerateRun).Run

	return &cmd, nil
}

func licenseGenerateRun(cmd *cobra.Command, args []string) error {
	mc, err := flagLicenseManager.Client(cmd)
	if err != nil {
		return err
	}

	did, err := flagDeploymentID.Get(cmd)
	if err != nil {
		return err
	}

	var inv *inventory.Spec

	if invFile, err := flagInventory.Get(cmd); err != nil {
		return err
	} else if invFile != "" {
		inv, err = ugrpc.UnmarshalFile[*inventory.Spec](invFile)
		if err != nil {
			return err
		}
	}

	l := logger.Str("ClusterID", did)

	l.Info("Generating License")

	lic, err := mc.License(cmd.Context(), lmanager.LicenseRequest{
		DeploymentID: util.NewType(did),
		Inventory:    util.NewType(ugrpc.NewObject(inv)),
	})
	if err != nil {
		return err
	}

	l = l.Str("LicenseID", lic.ID)

	l.Info("License Generated and printed to STDERR")

	fmt.Fprint(os.Stderr, lic.License)

	return nil
}
