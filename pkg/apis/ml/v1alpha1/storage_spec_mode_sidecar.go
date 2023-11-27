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
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	defaultRequestsCPU    = resource.MustParse("100m")
	defaultRequestsMemory = resource.MustParse("100Mi")
	defaultLimitsCPU      = resource.MustParse("200m")
	defaultLimitsMemory   = resource.MustParse("200Mi")
)

type ArangoMLStorageSpecModeSidecar struct {
	// ListenPort defines on which port the sidecar container will be listening for connections
	// +doc/default: 9201
	ListenPort *uint16 `json:"listenPort,omitempty"`

	// Resources holds resource requests & limits for container running the S3 proxy
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core
	Resources *core.ResourceRequirements `json:"resources,omitempty"`
}

func (s *ArangoMLStorageSpecModeSidecar) Validate() error {
	if s == nil {
		s = &ArangoMLStorageSpecModeSidecar{}
	}
	if s.GetListenPort() < 1 {
		return errors.Newf("invalid listenPort value: must be positive")
	}
	return nil
}

func (s *ArangoMLStorageSpecModeSidecar) GetListenPort() uint16 {
	if s == nil || s.ListenPort == nil {
		return 9201
	}
	return *s.ListenPort
}

func (s *ArangoMLStorageSpecModeSidecar) GetResources() core.ResourceRequirements {
	var resources core.ResourceRequirements
	if s != nil && s.Resources != nil {
		resources = *s.Resources
	}

	if len(resources.Requests) == 0 {
		resources.Requests = make(core.ResourceList)
		resources.Requests[core.ResourceCPU] = defaultRequestsCPU
		resources.Requests[core.ResourceMemory] = defaultRequestsMemory
	}
	if len(resources.Limits) == 0 {
		resources.Limits = make(core.ResourceList)
		resources.Limits[core.ResourceCPU] = defaultLimitsCPU
		resources.Limits[core.ResourceMemory] = defaultLimitsMemory
	}
	return resources
}
