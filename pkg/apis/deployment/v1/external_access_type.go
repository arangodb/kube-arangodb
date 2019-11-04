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

package v1

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// ExternalAccessType specifies the type of external access provides for the deployment
type ExternalAccessType string

const (
	// ExternalAccessTypeNone yields a cluster with no external access
	ExternalAccessTypeNone ExternalAccessType = "None"
	// ExternalAccessTypeAuto yields a cluster with an automatic selection for external access
	ExternalAccessTypeAuto ExternalAccessType = "Auto"
	// ExternalAccessTypeLoadBalancer yields a cluster with a service of type `LoadBalancer` to provide external access
	ExternalAccessTypeLoadBalancer ExternalAccessType = "LoadBalancer"
	// ExternalAccessTypeNodePort yields a cluster with a service of type `NodePort` to provide external access
	ExternalAccessTypeNodePort ExternalAccessType = "NodePort"
)

func (t ExternalAccessType) IsNone() bool         { return t == ExternalAccessTypeNone }
func (t ExternalAccessType) IsAuto() bool         { return t == ExternalAccessTypeAuto }
func (t ExternalAccessType) IsLoadBalancer() bool { return t == ExternalAccessTypeLoadBalancer }
func (t ExternalAccessType) IsNodePort() bool     { return t == ExternalAccessTypeNodePort }

// AsServiceType returns the k8s ServiceType for this ExternalAccessType.
// If type is "Auto", ServiceTypeLoadBalancer is returned.
func (t ExternalAccessType) AsServiceType() v1.ServiceType {
	switch t {
	case ExternalAccessTypeLoadBalancer, ExternalAccessTypeAuto:
		return v1.ServiceTypeLoadBalancer
	case ExternalAccessTypeNodePort:
		return v1.ServiceTypeNodePort
	default:
		return ""
	}
}

// Validate the type.
// Return errors when validation fails, nil on success.
func (t ExternalAccessType) Validate() error {
	switch t {
	case ExternalAccessTypeNone, ExternalAccessTypeAuto, ExternalAccessTypeLoadBalancer, ExternalAccessTypeNodePort:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown external access type: '%s'", string(t)))
	}
}

// NewExternalAccessType returns a reference to a string with given value.
func NewExternalAccessType(input ExternalAccessType) *ExternalAccessType {
	return &input
}

// NewExternalAccessTypeOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewExternalAccessTypeOrNil(input *ExternalAccessType) *ExternalAccessType {
	if input == nil {
		return nil
	}
	return NewExternalAccessType(*input)
}

// ExternalAccessTypeOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func ExternalAccessTypeOrDefault(input *ExternalAccessType, defaultValue ...ExternalAccessType) ExternalAccessType {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
