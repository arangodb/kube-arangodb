//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"

type DeploymentStatusHashes struct {
	Encryption DeploymentStatusHashesEncryption `json:"rocksDBEncryption,omitempty"`
	TLS        DeploymentStatusHashesTLS        `json:"tls,omitempty"`
	JWT        DeploymentStatusHashesJWT        `json:"jwt,omitempty"`
}

type DeploymentStatusHashesEncryption struct {
	Keys shared.HashList `json:"keys,omitempty"`

	Propagated bool `json:"propagated,omitempty"`
}

type DeploymentStatusHashesTLS struct {
	CA         *string         `json:"ca,omitempty"`
	Truststore shared.HashList `json:"truststore,omitempty"`

	Propagated bool `json:"propagated,omitempty"`
}

type DeploymentStatusHashesJWT struct {
	Active  string          `json:"active,omitempty"`
	Passive shared.HashList `json:"passive,omitempty"`

	Propagated bool `json:"propagated,omitempty"`
}
