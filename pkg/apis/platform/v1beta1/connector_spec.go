//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type ArangoPlatformConnectorSpec struct {
	// Description of what this connector does
	Description *string `json:"description,omitempty"`

	// Tags for discovery and filtering (e.g. "database", "aql", "vector-search")
	Tags []string `json:"tags,omitempty"`

	// Schema defines the JSON Schema for the connector's input query.
	// AI tools use this to validate parameters before submitting jobs.
	// Uses the same format as CRD validation schemas.
	// +doc/type: Object
	// +doc/link: Kubernetes JSON Schema docs|https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
	Schema *apiextensions.JSONSchemaProps `json:"schema,omitempty"`

	// Version of the connector
	Version *string `json:"version,omitempty"`
}

func (s *ArangoPlatformConnectorSpec) Validate() error {
	return nil
}
