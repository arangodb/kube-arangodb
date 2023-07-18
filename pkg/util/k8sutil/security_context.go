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

package k8sutil

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

// CreateSecurityContext returns security context.
// If secured container's feature is enabled then default values will set on nil fields.
func CreateSecurityContext(spec *api.ServerGroupSpecSecurityContext) *core.SecurityContext {
	return spec.NewSecurityContext(features.SecuredContainers().Enabled())
}

// CreatePodSecurityContext creates pod's security context.
func CreatePodSecurityContext(spec *api.ServerGroupSpecSecurityContext) *core.PodSecurityContext {
	return spec.NewPodSecurityContext(features.SecuredContainers().Enabled())
}
