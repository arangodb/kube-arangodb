//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"
	"path"

	"github.com/spf13/cobra"

	pbImplEnvoyConfigV1 "github.com/arangodb/kube-arangodb/integrations/envoy/config/v1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplEnvoyConfigV1.Name, func() Integration {
		return &envoyConfigV1{}
	})
}

type envoyConfigV1 struct {
}

func (a *envoyConfigV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return nil
}

func (a *envoyConfigV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	// Seed the initial ADS snapshot from the gateway CDS/LDS ConfigMap files mounted into the sidecar, so a
	// restarted gateway serves the last-known-good local config until the operator pushes an update.
	return pbImplEnvoyConfigV1.New(
		path.Join(utilConstants.GatewayCDSVolumeMountDir, utilConstants.GatewayConfigFileName),
		path.Join(utilConstants.GatewayLDSVolumeMountDir, utilConstants.GatewayConfigFileName),
		path.Join(utilConstants.GatewayCDSVolumeMountDir, utilConstants.GatewayConfigChecksum),
	)
}

func (a *envoyConfigV1) Name() string {
	return pbImplEnvoyConfigV1.Name
}

func (a *envoyConfigV1) Description() string {
	return "Enable EnvoyConfigV1 Integration Service"
}

func (*envoyConfigV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
