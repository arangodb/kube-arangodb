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

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ArangoPlatformConnectorType defines the connector pattern type
type ArangoPlatformConnectorType string

const (
	// ArangoPlatformConnectorTypeActive is the default connector type.
	// The connector actively polls for jobs and processes them.
	// +doc/enum: Active|Connector actively polls for and processes jobs (default)
	ArangoPlatformConnectorTypeActive ArangoPlatformConnectorType = "Active"
)

type ArangoPlatformConnectorSpec struct {
	// Type defines the connector execution pattern.
	// Currently only "Active" is supported — the connector runs as a long-lived process
	// that polls for pending jobs and processes them sequentially.
	// Set by the user when creating the connector. Defaults to "Active" if omitted.
	// +doc/default: Active
	// +doc/enum: Active|Connector actively polls for and processes jobs
	Type *ArangoPlatformConnectorType `json:"type,omitempty"`

	// Deployment is a reference to the ArangoDeployment that this connector belongs to.
	// Set by the user. Must point to an existing ArangoDeployment in the same namespace.
	// The operator verifies the deployment exists and sets the DeploymentFound condition.
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Route is a reference to the ArangoRoute that exposes this connector's external API
	// at a user-friendly path (e.g. /connector/<name>/). The route should redirect to
	// /_integration/connector/v1/ so that AI tools can use a clean per-connector URL.
	// Set by the user. The operator verifies the route exists and sets the RouteFound condition.
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Route *sharedApi.Object `json:"route,omitempty"`

	// Description is a human-readable text explaining what this connector does.
	// Shown to AI tools via /_inventory for discovery. Set by the user.
	// Example: "Execute AQL queries on ArangoDB"
	Description *string `json:"description,omitempty"`

	// Tags are labels used by AI tools to discover and filter connectors via /_inventory.
	// Set by the user. Use lowercase, descriptive terms.
	// Example: ["database", "aql", "query"]
	Tags []string `json:"tags,omitempty"`

	// Schema defines the JSON Schema that describes the expected format of the query
	// field when submitting jobs to this connector. AI tools read this from /_inventory
	// to validate input before creating a job. Set by the user.
	// Uses the standard Kubernetes JSONSchemaProps format (same as CRD validation schemas).
	// The platform validates submitted job queries against this schema.
	// +doc/type: Object
	// +doc/link: Kubernetes JSON Schema docs|https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
	Schema *apiextensions.JSONSchemaProps `json:"schema,omitempty"`

	// Version is the version string of the connector implementation.
	// Set by the user. Shown to AI tools via /_inventory. No format enforced,
	// but semantic versioning (e.g. "1.0.0") is recommended.
	Version *string `json:"version,omitempty"`
}

// GetType returns the connector type, defaulting to Active
func (s *ArangoPlatformConnectorSpec) GetType() ArangoPlatformConnectorType {
	if s == nil || s.Type == nil {
		return ArangoPlatformConnectorTypeActive
	}
	return *s.Type
}

func (s *ArangoPlatformConnectorSpec) Validate() error {
	if s == nil {
		return nil
	}

	switch s.GetType() {
	case ArangoPlatformConnectorTypeActive:
		return nil
	default:
		return errors.Errorf("unsupported connector type: %s", s.GetType())
	}
}
