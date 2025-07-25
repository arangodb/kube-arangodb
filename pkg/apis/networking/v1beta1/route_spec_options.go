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

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type ArangoRouteSpecOptions struct {
	// Upgrade keeps the connection upgrade options
	Upgrade ArangoRouteSpecOptionsUpgrade `json:"upgrade,omitempty"`
}

func (a *ArangoRouteSpecOptions) AsStatus() *ArangoRouteStatusTargetOptions {
	if a == nil {
		return nil
	}

	return &ArangoRouteStatusTargetOptions{
		Upgrade: a.Upgrade.asStatus(),
	}
}

func (a *ArangoRouteSpecOptions) Validate() error {
	if a == nil {
		a = &ArangoRouteSpecOptions{}
	}

	if err := shared.WithErrors(
		shared.ValidateOptionalInterfacePath("upgrade", a.Upgrade),
	); err != nil {
		return err
	}

	return nil
}
