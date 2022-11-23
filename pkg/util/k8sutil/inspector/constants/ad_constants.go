//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deploymentv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

// ArangoDeployment
const (
	ArangoDeploymentGroup     = deployment.ArangoDeploymentGroupName
	ArangoDeploymentResource  = deployment.ArangoDeploymentResourcePlural
	ArangoDeploymentKind      = deployment.ArangoDeploymentResourceKind
	ArangoDeploymentVersionV1 = deploymentv1.ArangoDeploymentVersion
)

func ArangoDeploymentGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoDeploymentGroup,
		Kind:  ArangoDeploymentKind,
	}
}

func ArangoDeploymentGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoDeploymentGroup,
		Kind:    ArangoDeploymentKind,
		Version: ArangoDeploymentVersionV1,
	}
}

func ArangoDeploymentGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoDeploymentGroup,
		Resource: ArangoDeploymentResource,
	}
}

func ArangoDeploymentGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoDeploymentGroup,
		Resource: ArangoDeploymentResource,
		Version:  ArangoDeploymentVersionV1,
	}
}
