//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"

	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sidecarSvcAuthzClient "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/client"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/integration"
)

type ConfigurationType string

func (c ConfigurationType) Validate() error {
	switch c {
	case ConfigurationTypeAlways, ConfigurationTypeCentral, ConfigurationTypeCentralPermissive, ConfigurationTypeNever:
		return nil
	default:
		return errors.Errorf("Configuration type '%s' is not supported", string(c))
	}
}

func (c ConfigurationType) String() string {
	return string(c)
}

const (
	ConfigurationTypeNever             ConfigurationType = "never"
	ConfigurationTypeAlways            ConfigurationType = "always"
	ConfigurationTypeCentral           ConfigurationType = "central"
	ConfigurationTypeCentralPermissive ConfigurationType = "central-permissive"
)

func NewConfiguration() Configuration {
	return Configuration{}
}

type Configuration struct {
	Type ConfigurationType `json:"type,omitempty"`
}

func (c Configuration) Plugin(ctx context.Context) (pbImplAuthorizationV1Shared.Plugin, error) {
	switch c.Type {
	case ConfigurationTypeNever:
		return pbImplAuthorizationV1Shared.NewNeverPlugin(), nil
	case ConfigurationTypeAlways:
		return pbImplAuthorizationV1Shared.NewAlwaysPlugin(), nil
	case ConfigurationTypeCentral:
		return pbImplAuthorizationV1Shared.SuperUser(sidecarSvcAuthzClient.NewClient(ctx, integration.NewIntegrationClientCache(sidecarSvcAuthzDefinition.NewAuthorizationPoolServiceClient))), nil

	case ConfigurationTypeCentralPermissive:
		return pbImplAuthorizationV1Shared.SuperUser(pbImplAuthorizationV1Shared.Permissive(sidecarSvcAuthzClient.NewClient(ctx, integration.NewIntegrationClientCache(sidecarSvcAuthzDefinition.NewAuthorizationPoolServiceClient)), logger)), nil
	default:
		return nil, errors.Errorf("Configuration type '%s' is not supported", string(c.Type))
	}
}

func (c Configuration) Validate() error {
	return errors.Errors(
		shared.PrefixResourceError("type", c.Type.Validate()),
	)
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
