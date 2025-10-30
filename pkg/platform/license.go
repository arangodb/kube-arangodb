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
	goHttp "net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/platform/inventory"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func license() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "license"
	cmd.Short = "License related Operations"

	if err := cli.RegisterFlags(&cmd); err != nil {
		return nil, err
	}

	if err := withRegisterCommand(&cmd,
		licenseInventory,
		licenseSecret,
		licenseActivate,
		licenseGenerate,
	); err != nil {
		return nil, err
	}

	return &cmd, nil
}

func buildInventory(cmd *cobra.Command) (*inventory.Spec, error) {
	logger.Info("Connecting to the server...")

	conn, err := flagDeployment.Connection(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := arangod.GetRequestWithTimeout[driver.VersionInfo](cmd.Context(), globals.GetGlobals().Timeouts().ArangoD().Get(), conn, "_api", "version").
		AcceptCode(goHttp.StatusOK).
		Response()
	if err != nil {
		return nil, err
	}

	var cfg inventory.Configuration

	if v, err := flagTelemetry.Get(cmd); !v || err != nil {
		cfg.Telemetry = util.NewType(false)
	}

	logger.Info("Discovered Arango %s (%s)", resp.Version, resp.License)

	obj, err := inventory.FetchInventory(cmd.Context(), logger, 8, conn, &cfg)

	if err != nil {
		return nil, err
	}

	obj = util.FilterList(obj, func(item *inventory.Item) bool {
		return item != nil
	})

	did := util.FilterList(obj, util.MultiFilterList(
		func(item *inventory.Item) bool {
			return item.Type == "ARANGO_DEPLOYMENT"
		},
		func(item *inventory.Item) bool {
			v, ok := item.Dimensions["detail"]
			return ok && v == "id"
		},
	))

	if len(did) != 1 {
		return nil, errors.Errorf("Expected to find a single ARANGO_DEPLOYMENT ID")
	}

	tz, err := did[0].GetValue().Type()
	if err != nil {
		return nil, err
	}

	if tz != reflect.TypeFor[string]() {
		return nil, errors.Errorf("Expected to find type for ARANGO_DEPLOYMENT ID")
	}

	return &inventory.Spec{
		DeploymentId: did[0].GetValue().GetStr(),
		Items:        obj,
	}, nil
}
