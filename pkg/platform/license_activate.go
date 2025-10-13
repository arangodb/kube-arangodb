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
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/license/manager"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

func licenseActivate() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "activate"
	cmd.Short = "Activates the License on ArangoDB Endpoint"

	if err := cli.RegisterFlags(&cmd, flagLicenseManagerEndpoint, flagLicenseManagerClientID, flagLicenseManagerClientSecret, flagActivateInterval, flagDeployment); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(licenseActivateRun).Run

	return &cmd, nil
}

func licenseActivateRun(cmd *cobra.Command, args []string) error {
	mc, err := getManagerClient(cmd)
	if err != nil {
		return err
	}

	conn, err := flagDeployment.Connection(cmd)
	if err != nil {
		return err
	}

	c := client.NewClient(conn, nil)

	del, err := flagActivateInterval.Get(cmd)
	if err != nil {
		return err
	}

	if del == 0 {
		logger.Info("Activate Once")

		return licenseActivateExecute(cmd.Context(), logger, c, mc)
	}

	intervalT := time.NewTicker(del)
	defer intervalT.Stop()

	logger.Dur("interval", del).Info("Activate In interval")

	for {
		if err := licenseActivateExecute(cmd.Context(), logger, c, mc); err != nil {
			return err
		}

		select {
		case <-intervalT.C:
			continue
		case <-cmd.Context().Done():
			return nil
		}
	}
}

func licenseActivateExecute(ctx context.Context, logger logging.Logger, c client.Client, mc manager.Client) error {
	logger.Info("Connecting to server...")
	id, err := c.DeploymentID(ctx)
	if err != nil {
		return err
	}

	l := logger.Str("ClusterID", id.Id)

	l.Info("Discovered ClusterID")

	l.Info("Generating License")

	lic, err := mc.License(ctx, manager.LicenseRequest{
		DeploymentID: util.NewType(id.Id),
	})
	if err != nil {
		return err
	}

	l = l.Str("LicenseID", lic.ID)

	l.Info("Activating license...")

	if err := c.SetLicense(ctx, lic.License, true); err != nil {
		return err
	}

	nlic, err := c.GetLicense(ctx)
	if err != nil {
		return err
	}

	l.Str("hash", nlic.Hash).Info("Activated!")

	return nil
}
