//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

import (
	"net"
	"net/url"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ExternalAccessSpec holds configuration for the external access provided for the deployment.
type ExternalAccessSpec struct {

	// Type specifies the type of Service that will be created to provide access to the ArangoDB deployment from outside the Kubernetes cluster.
	// +doc/enum: Auto|Create a Service of type LoadBalancer and fallback to a Service or type NodePort when the LoadBalancer is not assigned an IP address.
	// +doc/enum: None|limit access to application running inside the Kubernetes cluster.
	// +doc/enum: LoadBalancer|Create a Service of type LoadBalancer for the ArangoDB deployment.
	// +doc/enum: NodePort|Create a Service of type NodePort for the ArangoDB deployment.
	// +doc/enum: Managed|Manages only existing services.
	Type *ExternalAccessType `json:"type,omitempty"`

	// NodePort define optional port used in case of Auto or NodePort type.
	// This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.
	// If you do not specify this setting, a random port will be chosen automatically.
	NodePort *int `json:"nodePort,omitempty"`

	// LoadBalancerIP define optional IP used to configure a load-balancer on, in case of Auto or LoadBalancer type.
	// If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`

	// LoadBalancerSourceRanges define LoadBalancerSourceRanges used for LoadBalancer Service type
	// If specified and supported by the platform, this will restrict traffic through the cloud-provider
	// load-balancer will be restricted to the specified client IPs. This field will be ignored if the
	// cloud-provider does not support the feature.
	// +doc/link: Cloud Provider Firewall|https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty"`

	// AdvertisedEndpoint is passed to the coordinators/single servers for advertising a specific endpoint
	AdvertisedEndpoint *string `json:"advertisedEndpoint,omitempty"`

	// ManagedServiceNames keeps names of services which are not managed by KubeArangoDB.
	// It is only relevant when type of service is `managed`.
	ManagedServiceNames []string `json:"managedServiceNames,omitempty"`
}

// GetType returns the value of type.
func (s ExternalAccessSpec) GetType() ExternalAccessType {
	return ExternalAccessTypeOrDefault(s.Type, ExternalAccessTypeAuto)
}

// GetNodePort returns the value of nodePort.
func (s ExternalAccessSpec) GetNodePort() int {
	return util.TypeOrDefault[int](s.NodePort)
}

// GetLoadBalancerIP returns the value of loadBalancerIP.
func (s ExternalAccessSpec) GetLoadBalancerIP() string {
	return util.TypeOrDefault[string](s.LoadBalancerIP)
}

// GetAdvertisedEndpoint returns the advertised endpoint or empty string if none was specified
func (s ExternalAccessSpec) GetAdvertisedEndpoint() string {
	return util.TypeOrDefault[string](s.AdvertisedEndpoint)
}

// GetManagedServiceNames returns a list of managed service names.
func (s ExternalAccessSpec) GetManagedServiceNames() []string {
	return s.ManagedServiceNames
}

// HasAdvertisedEndpoint return whether an advertised endpoint was specified or not
func (s ExternalAccessSpec) HasAdvertisedEndpoint() bool {
	return s.AdvertisedEndpoint != nil
}

// Validate the given spec
func (s ExternalAccessSpec) Validate() error {
	if err := s.GetType().Validate(); err != nil {
		return errors.WithStack(err)
	}
	if s.AdvertisedEndpoint != nil {
		ep := s.GetAdvertisedEndpoint()
		if _, err := url.Parse(ep); err != nil {
			return errors.WithStack(errors.Errorf("Failed to parse advertised endpoint '%s': %s", ep, err))
		}
	}
	for _, x := range s.LoadBalancerSourceRanges {
		if _, _, err := net.ParseCIDR(x); err != nil {
			return errors.WithStack(errors.Errorf("Failed to parse loadbalancer source range '%s': %s", x, err))
		}
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
		s.NodePort = util.NewTypeOrNil[int](source.NodePort)
	}
	if s.LoadBalancerIP == nil {
		s.LoadBalancerIP = util.NewTypeOrNil[string](source.LoadBalancerIP)
	}
	if s.LoadBalancerSourceRanges == nil && len(source.LoadBalancerSourceRanges) > 0 {
		s.LoadBalancerSourceRanges = append([]string{}, source.LoadBalancerSourceRanges...)
	}
	if s.AdvertisedEndpoint == nil {
		s.AdvertisedEndpoint = source.AdvertisedEndpoint
	}
	if s.ManagedServiceNames == nil && len(source.ManagedServiceNames) > 0 {
		s.ManagedServiceNames = append([]string{}, source.ManagedServiceNames...)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s ExternalAccessSpec) ResetImmutableFields(fieldPrefix string, target *ExternalAccessSpec) []string {
	return nil
}
