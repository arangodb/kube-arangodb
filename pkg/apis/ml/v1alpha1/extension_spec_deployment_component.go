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

package v1alpha1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLExtensionSpecDeploymentComponent struct {
	// Port defines on which port the container will be listening for connections
	Port *int32 `json:"port,omitempty"`

	// ServiceType determines how the Service is exposed
	// +doc/default: ClusterIP
	// +doc/link: Kubernetes Documentation|https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	ServiceType *core.ServiceType `json:"serviceType,omitempty"`

	// Image defines image used for the component
	*sharedApi.Image `json:",inline"`

	// Resources holds resource requests & limits for container
	// If not specified, default values will be used
	*sharedApi.Resources `json:",inline"`
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetPort() int32 {
	if s == nil || s.Port == nil {
		return 0
	}
	return *s.Port
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetImage() *sharedApi.Image {
	if s == nil || s.Image == nil {
		return nil
	}

	return s.Image
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetResources() *sharedApi.Resources {
	if s == nil || s.Resources == nil {
		return nil
	}

	return s.Resources
}

func (s *ArangoMLExtensionSpecDeploymentComponent) GetServiceType() core.ServiceType {
	if s == nil || s.ServiceType == nil {
		return core.ServiceTypeClusterIP
	}

	return *s.ServiceType
}

func (s *ArangoMLExtensionSpecDeploymentComponent) Validate() error {
	if s == nil {
		return nil
	}

	var err []error

	if s.GetPort() < 1 {
		err = append(err, shared.PrefixResourceErrors("port", errors.Newf("must be positive")))
	}

	err = append(err,
		shared.PrefixResourceErrors("resources", s.GetResources().Validate()),
		shared.PrefixResourceErrors("image", shared.ValidateRequired(s.GetImage(), func(obj sharedApi.Image) error { return obj.Validate() })),
		shared.PrefixResourceErrors("image", shared.ValidateServiceType(s.GetServiceType())),
	)

	return shared.WithErrors(err...)
}
