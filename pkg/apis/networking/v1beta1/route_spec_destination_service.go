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
	"k8s.io/apimachinery/pkg/util/intstr"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoRouteSpecDestinationService struct {
	// +doc/skip: uid
	// +doc/skip: checksum
	*sharedApi.Object `json:",inline,omitempty"`

	// Port defines Port or Port Name used as destination
	// +doc/type: intstr.IntOrString
	// +doc/required
	Port *intstr.IntOrString `json:"port,omitempty"`

	// Mode defiles the resolve mode for the service discovery
	// +doc/default: dns
	// +doc/enum: dns|DNS Names of Service used
	// +doc/enum: ip|IP used wherever possible (except Headless Services)
	Mode *ArangoRouteSpecResolveMode `json:"mode,omitempty"`
}

func (a *ArangoRouteSpecDestinationService) GetPort() *intstr.IntOrString {
	if a == nil || a.Port == nil {
		return nil
	}

	return a.Port
}

func (a *ArangoRouteSpecDestinationService) Validate() error {
	if a == nil {
		a = &ArangoRouteSpecDestinationService{}
	}

	if err := shared.WithErrors(a.Object.Validate(), shared.ValidateRequiredPath("port", a.Port, func(i intstr.IntOrString) error {
		return nil
	})); err != nil {
		return err
	}

	return nil
}
