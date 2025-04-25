//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
)

// ArangoSchedulerDeployment
const (
	ArangoSchedulerDeploymentGroup          = scheduler.ArangoSchedulerGroupName
	ArangoSchedulerDeploymentResource       = scheduler.DeploymentResourcePlural
	ArangoSchedulerDeploymentKind           = scheduler.DeploymentResourceKind
	ArangoSchedulerDeploymentVersionV1Beta1 = schedulerApi.ArangoSchedulerVersion
)

func init() {
	register[*schedulerApi.ArangoSchedulerDeployment](ArangoSchedulerDeploymentGKv1Beta1(), ArangoSchedulerDeploymentGRv1Beta1())
}

func ArangoSchedulerDeploymentGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoSchedulerDeploymentGroup,
		Kind:  ArangoSchedulerDeploymentKind,
	}
}

func ArangoSchedulerDeploymentGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoSchedulerDeploymentGroup,
		Kind:    ArangoSchedulerDeploymentKind,
		Version: ArangoSchedulerDeploymentVersionV1Beta1,
	}
}

func ArangoSchedulerDeploymentGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoSchedulerDeploymentGroup,
		Resource: ArangoSchedulerDeploymentResource,
	}
}

func ArangoSchedulerDeploymentGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoSchedulerDeploymentGroup,
		Resource: ArangoSchedulerDeploymentResource,
		Version:  ArangoSchedulerDeploymentVersionV1Beta1,
	}
}
