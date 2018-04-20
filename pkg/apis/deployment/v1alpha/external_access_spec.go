//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// ExternalAccessSpec holds configuration for the external access provided for the deployment.
type ExternalAccessSpec struct {
	// Type of external access
	Type *ExternalAccessType `json:"type,omitempty"`
	// Optional port used in case of Auto or NodePort type.
	NodePort *int `json:"nodePort,omitempty"`
	// Optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`
}

// GetType returns the value of type.
func (s ExternalAccessSpec) GetType() ExternalAccessType {
	return ExternalAccessTypeOrDefault(s.Type, ExternalAccessTypeAuto)
}

// GetNodePort returns the value of nodePort.
func (s ExternalAccessSpec) GetNodePort() int {
	return util.IntOrDefault(s.NodePort)
}

// GetLoadBalancerIP returns the value of loadBalancerIP.
func (s ExternalAccessSpec) GetLoadBalancerIP() string {
	return util.StringOrDefault(s.LoadBalancerIP)
}

// Validate the given spec
func (s ExternalAccessSpec) Validate() error {
	if err := s.GetType().Validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *ExternalAccessSpec) SetDefaults() {
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *ExternalAccessSpec) SetDefaultsFrom(source ExternalAccessSpec) {
	if s.Type == nil {
		s.Type = NewExternalAccessTypeOrNil(source.Type)
	}
	if s.NodePort == nil {
		s.NodePort = util.NewIntOrNil(source.NodePort)
	}
	if s.LoadBalancerIP == nil {
		s.LoadBalancerIP = util.NewStringOrNil(source.LoadBalancerIP)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s ExternalAccessSpec) ResetImmutableFields(fieldPrefix string, target *ExternalAccessSpec) []string {
	return nil
}
