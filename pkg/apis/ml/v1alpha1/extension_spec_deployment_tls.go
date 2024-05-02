//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

type ArangoMLExtensionSpecDeploymentTLS struct {
	// Enabled define if TLS Should be enabled. If is not set then default is taken from ArangoDeployment settings
	Enabled *bool `json:"enabled,omitempty"`

	// AltNames define TLS AltNames used when TLS on the ArangoDB is enabled
	AltNames []string `json:"altNames,omitempty"`
}

func (a *ArangoMLExtensionSpecDeploymentTLS) IsEnabled() bool {
	if a == nil || a.Enabled == nil {
		return true
	}

	return *a.Enabled
}

func (a *ArangoMLExtensionSpecDeploymentTLS) GetAltNames() []string {
	if a == nil || a.AltNames == nil {
		return nil
	}

	return a.AltNames
}
