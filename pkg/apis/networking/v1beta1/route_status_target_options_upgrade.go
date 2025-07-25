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
// See the License for the Statusific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package v1beta1

import "github.com/arangodb/kube-arangodb/pkg/util"

type ArangoRouteStatusTargetOptionsUpgrade []ArangoRouteStatusTargetOptionUpgrade

func (a ArangoRouteStatusTargetOptionsUpgrade) Hash() string {
	if len(a) == 0 {
		return ""
	}

	return util.SHA256FromStringArray(util.FormatList(a, func(a ArangoRouteStatusTargetOptionUpgrade) string {
		return a.Hash()
	})...)
}

type ArangoRouteStatusTargetOptionUpgrade struct {
	// Type defines type of the Upgrade
	// +doc/enum: websocket|HTTP WebSocket Upgrade type
	Type ArangoRouteUpgradeOptionType `json:"type"`

	// Enabled defines if upgrade option is enabled
	Enabled *bool `json:"enabled,omitempty"`
}

func (a *ArangoRouteStatusTargetOptionUpgrade) Hash() string {
	if a == nil {
		return ""
	}

	return util.SHA256FromStringArray(string(a.Type), util.BoolSwitch(util.WithDefault(a.Enabled), "true", "false"))
}
