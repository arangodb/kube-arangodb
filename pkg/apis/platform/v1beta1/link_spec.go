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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ArangoPlatformLinkType defines the link pattern type
type ArangoPlatformLinkType string

const (
	// ArangoPlatformLinkTypeActive is the default link type.
	// The link actively polls for jobs and processes them.
	// +doc/enum: Active|Link actively polls for and processes jobs (default)
	ArangoPlatformLinkTypeActive ArangoPlatformLinkType = "Active"
)

type ArangoPlatformLinkSpec struct {
	// Type defines the link execution pattern.
	// Currently only "Active" is supported — the link runs as a long-lived process
	// that polls for pending jobs and processes them sequentially.
	// Set by the user when creating the link. Defaults to "Active" if omitted.
	// +doc/default: Active
	// +doc/enum: Active|Link actively polls for and processes jobs
	Type *ArangoPlatformLinkType `json:"type,omitempty"`

	// Deployment is a reference to the ArangoDeployment that this link belongs to.
	// Set by the user. Must point to an existing ArangoDeployment in the same namespace.
	// The operator verifies the deployment exists and sets the DeploymentFound condition.
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Deployment *sharedApi.Object `json:"deployment,omitempty"`

	// Route is a reference to the ArangoRoute that exposes this link's external API
	// at a user-friendly path (e.g. /link/<name>/). The route should redirect to
	// /_integration/link/v1/ so that AI tools can use a clean per-link URL.
	// Set by the user. The operator verifies the route exists and sets the RouteFound condition.
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	Route *sharedApi.Object `json:"route,omitempty"`
}

// GetType returns the link type, defaulting to Active
func (s *ArangoPlatformLinkSpec) GetType() ArangoPlatformLinkType {
	if s == nil || s.Type == nil {
		return ArangoPlatformLinkTypeActive
	}
	return *s.Type
}

func (s *ArangoPlatformLinkSpec) Validate() error {
	if s == nil {
		return nil
	}

	switch s.GetType() {
	case ArangoPlatformLinkTypeActive:
		return nil
	default:
		return errors.Errorf("unsupported link type: %s", s.GetType())
	}
}
