//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoRouteSpecOptionsUpgrade []ArangoRouteSpecOptionUpgrade

func (a ArangoRouteSpecOptionsUpgrade) Validate() error {
	return shared.ValidateInterfaceList(a)
}

func (a ArangoRouteSpecOptionsUpgrade) asStatus() ArangoRouteStatusTargetOptionsUpgrade {
	if len(a) == 0 {
		return nil
	}

	return util.FormatList(a, func(a ArangoRouteSpecOptionUpgrade) ArangoRouteStatusTargetOptionUpgrade {
		return a.asStatus()
	})
}

type ArangoRouteSpecOptionUpgrade struct {
	// Type defines type of the Upgrade
	// +doc/enum: websocket|HTTP WebSocket Upgrade type
	Type ArangoRouteUpgradeOptionType `json:"type"`

	// Enabled defines if upgrade option is enabled
	Enabled *bool `json:"enabled,omitempty"`
}

func (a ArangoRouteSpecOptionUpgrade) asStatus() ArangoRouteStatusTargetOptionUpgrade {
	return ArangoRouteStatusTargetOptionUpgrade{
		Type:    a.Type,
		Enabled: util.NewType(util.WithDefault(a.Enabled)),
	}
}

func (a ArangoRouteSpecOptionUpgrade) Validate() error {
	if err := shared.WithErrors(
		shared.ValidateRequiredInterfacePath("type", a.Type),
	); err != nil {
		return err
	}

	return nil
}
