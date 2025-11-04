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

package v1

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

// DeploymentStatusLicense contains the status part of a Cluster resource.
type DeploymentStatusLicense struct {
	// ID Defines the License ID
	ID string `json:"id,omitempty"`

	// Hash Defines the License Hash
	Hash string `json:"hash,omitempty"`

	// Expires Defines the expiration time of the License
	Expires meta.Time `json:"expires,omitempty"`

	// Regenerate Defines the time when license will be regenerated
	Regenerate meta.Time `json:"regenerate,omitempty"`

	// Mode defines the license mode
	Mode LicenseMode `json:"mode,omitempty"`
}

// Equal checks for equality
func (ds *DeploymentStatusLicense) Equal(other *DeploymentStatusLicense) bool {
	if ds == nil && other == nil {
		return true
	}

	if ds == nil || other == nil {
		return false
	}

	return ds.ID == other.ID && ds.Hash == other.Hash
}
