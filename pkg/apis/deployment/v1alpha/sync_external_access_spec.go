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
	"fmt"
	"net"
	"net/url"
	"strconv"
)

// SyncExternalAccessSpec holds configuration for the external access provided for the sync deployment.
type SyncExternalAccessSpec struct {
	ExternalAccessSpec
	MasterEndpoint []string `json:"masterEndpoint,omitempty"`
}

// GetMasterEndpoint returns the value of masterEndpoint.
func (s SyncExternalAccessSpec) GetMasterEndpoint() []string {
	return s.MasterEndpoint
}

// ResolveMasterEndpoint returns the value of `--master.endpoint` option passed to arangosync.
func (s SyncExternalAccessSpec) ResolveMasterEndpoint(syncServiceHostName string, syncServicePort int) []string {
	if len(s.MasterEndpoint) > 0 {
		return s.MasterEndpoint
	}
	if ip := s.GetLoadBalancerIP(); ip != "" {
		syncServiceHostName = ip
	}
	return []string{"https://" + net.JoinHostPort(syncServiceHostName, strconv.Itoa(syncServicePort))}
}

// Validate the given spec
func (s SyncExternalAccessSpec) Validate() error {
	if err := s.ExternalAccessSpec.Validate(); err != nil {
		return maskAny(err)
	}
	for _, ep := range s.MasterEndpoint {
		if _, err := url.Parse(ep); err != nil {
			return maskAny(fmt.Errorf("Failed to parse master endpoint '%s': %s", ep, err))
		}
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SyncExternalAccessSpec) SetDefaults() {
	s.ExternalAccessSpec.SetDefaults()
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *SyncExternalAccessSpec) SetDefaultsFrom(source SyncExternalAccessSpec) {
	s.ExternalAccessSpec.SetDefaultsFrom(source.ExternalAccessSpec)
	if s.MasterEndpoint == nil && source.MasterEndpoint != nil {
		s.MasterEndpoint = append([]string{}, source.MasterEndpoint...)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s SyncExternalAccessSpec) ResetImmutableFields(fieldPrefix string, target *SyncExternalAccessSpec) []string {
	result := s.ExternalAccessSpec.ResetImmutableFields(fieldPrefix, &s.ExternalAccessSpec)
	return result
}
