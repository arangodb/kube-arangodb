//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import core "k8s.io/api/core/v1"

// DeploymentCommunicationMethod define communication method used for inter-cluster communication
type DeploymentCommunicationMethod string

const (
	// DefaultDeploymentCommunicationMethod define default communication method.
	DefaultDeploymentCommunicationMethod = DeploymentCommunicationMethodHeadlessService
	// DeploymentCommunicationMethodHeadlessService define old communication mechanism, based on headless service.
	DeploymentCommunicationMethodHeadlessService DeploymentCommunicationMethod = "headless"
	// DeploymentCommunicationMethodDNS define ClusterIP Service DNS based communication.
	DeploymentCommunicationMethodDNS DeploymentCommunicationMethod = "dns"
	// DeploymentCommunicationMethodShortDNS define ClusterIP Service DNS based communication. Use namespaced short DNS (used in migration)
	DeploymentCommunicationMethodShortDNS DeploymentCommunicationMethod = "short-dns"
	// DeploymentCommunicationMethodHeadlessDNS define Headless Service DNS based communication.
	DeploymentCommunicationMethodHeadlessDNS DeploymentCommunicationMethod = "headless-dns"
	// DeploymentCommunicationMethodIP define ClusterIP Service IP based communication.
	DeploymentCommunicationMethodIP DeploymentCommunicationMethod = "ip"
)

// Get returns communication method from pointer. If pointer is nil default is returned.
func (d *DeploymentCommunicationMethod) Get() DeploymentCommunicationMethod {
	if d == nil {
		return DefaultDeploymentCommunicationMethod
	}

	switch v := *d; v {
	case DeploymentCommunicationMethodHeadlessService, DeploymentCommunicationMethodDNS, DeploymentCommunicationMethodIP, DeploymentCommunicationMethodShortDNS, DeploymentCommunicationMethodHeadlessDNS:
		return v
	default:
		return DefaultDeploymentCommunicationMethod
	}
}

// ServiceType returns Service Type for communication method
func (d *DeploymentCommunicationMethod) ServiceType() core.ServiceType {
	switch d.Get() {
	default:
		return core.ServiceTypeClusterIP
	}
}

// ServiceClusterIP returns Service ClusterIP for communication method
func (d *DeploymentCommunicationMethod) ServiceClusterIP() string {
	switch d.Get() {
	case DeploymentCommunicationMethodHeadlessDNS:
		return core.ClusterIPNone
	default:
		return ""
	}
}

// String returns string representation of method.
func (d DeploymentCommunicationMethod) String() string {
	return string(d)
}

// New returns pointer.
func (d DeploymentCommunicationMethod) New() *DeploymentCommunicationMethod {
	return &d
}
