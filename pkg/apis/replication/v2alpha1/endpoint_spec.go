//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"net/url"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// EndpointSpec contains the specification used to reach the syncmasters
// in either source or destination mode.
type EndpointSpec struct {
	// DeploymentName holds the name of an ArangoDeployment resource.
	// If set this provides default values for masterEndpoint, auth & tls.
	DeploymentName *string `json:"deploymentName,omitempty"`
	// MasterEndpoint holds a list of URLs used to reach the syncmaster(s).
	MasterEndpoint []string `json:"masterEndpoint,omitempty"`
	// Authentication holds settings needed to authentication at the syncmaster.
	Authentication EndpointAuthenticationSpec `json:"auth"`
	// TLS holds settings needed to verify the TLS connection to the syncmaster.
	TLS EndpointTLSSpec `json:"tls"`
}

// GetDeploymentName returns the value of deploymentName.
func (s EndpointSpec) GetDeploymentName() string {
	return util.TypeOrDefault[string](s.DeploymentName)
}

// HasDeploymentName returns the true when a non-empty deployment name it set.
func (s EndpointSpec) HasDeploymentName() bool {
	return s.GetDeploymentName() != ""
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s EndpointSpec) Validate(isSourceEndpoint bool) error {
	if err := shared.ValidateOptionalResourceName(s.GetDeploymentName()); err != nil {
		return errors.WithStack(err)
	}
	for _, ep := range s.MasterEndpoint {
		if _, err := url.Parse(ep); err != nil {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid master endpoint '%s': %s", ep, err))
		}
	}
	hasDeploymentName := s.HasDeploymentName()
	if !hasDeploymentName && len(s.MasterEndpoint) == 0 {
		return errors.WithStack(errors.Wrapf(ValidationError, "Provide a deploy name or at least one master endpoint"))
	}
	if err := s.Authentication.Validate(isSourceEndpoint || !hasDeploymentName); err != nil {
		return errors.WithStack(err)
	}
	if err := s.TLS.Validate(!hasDeploymentName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *EndpointSpec) SetDefaults() {
	s.Authentication.SetDefaults()
	s.TLS.SetDefaults()
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *EndpointSpec) SetDefaultsFrom(source EndpointSpec) {
	if s.DeploymentName == nil {
		s.DeploymentName = util.NewTypeOrNil[string](source.DeploymentName)
	}
	s.Authentication.SetDefaultsFrom(source.Authentication)
	s.TLS.SetDefaultsFrom(source.TLS)
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s EndpointSpec) ResetImmutableFields(target *EndpointSpec, fieldPrefix string) []string {
	var result []string
	if s.GetDeploymentName() != target.GetDeploymentName() {
		result = append(result, fieldPrefix+"deploymentName")
	}
	if list := s.Authentication.ResetImmutableFields(&target.Authentication, fieldPrefix+"auth."); len(list) > 0 {
		result = append(result, list...)
	}
	if list := s.TLS.ResetImmutableFields(&target.TLS, fieldPrefix+"tls."); len(list) > 0 {
		result = append(result, list...)
	}
	return result
}
