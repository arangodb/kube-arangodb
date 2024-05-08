//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type GraphAnalyticsEngineSpecDeploymentService struct {
	// Type determines how the Service is exposed
	// +doc/enum: ClusterIP|service will only be accessible inside the cluster, via the cluster IP
	// +doc/enum: NodePort|service will be exposed on one port of every node, in addition to 'ClusterIP' type
	// +doc/enum: LoadBalancer|service will be exposed via an external load balancer (if the cloud provider supports it), in addition to 'NodePort' type
	// +doc/enum: ExternalName|service consists of only a reference to an external name that kubedns or equivalent will return as a CNAME record, with no exposing or proxying of any pods involved
	// +doc/enum: None|service is not created
	// +doc/default: ClusterIP
	// +doc/link: Kubernetes Documentation|https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	Type *core.ServiceType `json:"type,omitempty"`
}

func (g *GraphAnalyticsEngineSpecDeploymentService) GetType() core.ServiceType {
	if g == nil || g.Type == nil {
		return core.ServiceTypeClusterIP
	}

	return *g.Type
}

func (g *GraphAnalyticsEngineSpecDeploymentService) Validate() error {
	if g == nil {
		return nil
	}

	errs := []error{
		shared.PrefixResourceErrors("type", shared.ValidateServiceType(g.GetType())),
	}
	return shared.WithErrors(errs...)
}
