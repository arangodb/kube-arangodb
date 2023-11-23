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
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ArangoMLStorageSpec struct {
	// ListenPort defines on which port the sidecar container will be listening for connections
	// +doc/default: 9201
	ListenPort *uint16 `json:"listenPort,omitempty"`

	// Resources holds resource requests & limits for container running the S3 proxy
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core
	Resources core.ResourceRequirements `json:"resources,omitempty"`

	S3 *ArangoMLStorageS3Spec `json:"s3,omitempty"`
}

func (s *ArangoMLStorageSpec) Validate() error {
	if s.S3 == nil {
		return errors.New("Currently only s3 storage type is supported")
	}

	return s.S3.Validate()
}

// SetDefaults fills in missing defaults
func (s *ArangoMLStorageSpec) SetDefaults() {
	if s == nil {
		return
	}

	resources := s.Resources
	if len(resources.Requests) == 0 {
		resources.Requests = make(core.ResourceList)
		resources.Requests[core.ResourceCPU] = resource.MustParse("100m")
		resources.Requests[core.ResourceMemory] = resource.MustParse("100m")
	}
	if len(resources.Limits) == 0 {
		resources.Limits = make(core.ResourceList)
		resources.Limits[core.ResourceCPU] = resource.MustParse("250m")
		resources.Limits[core.ResourceMemory] = resource.MustParse("250m")
	}
}
