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

	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
)

// ArangoMLExtension
const (
	ArangoMLExtensionGroup           = ml.ArangoMLGroupName
	ArangoMLExtensionResource        = ml.ArangoMLExtensionResourcePlural
	ArangoMLExtensionKind            = ml.ArangoMLExtensionResourceKind
	ArangoMLExtensionVersionV1Alpha1 = mlApiv1alpha1.ArangoMLVersion
	ArangoMLExtensionVersionV1Beta1  = mlApi.ArangoMLVersion
)

func init() {
	register[*mlApiv1alpha1.ArangoMLExtension](ArangoMLExtensionGKv1Alpha1(), ArangoMLExtensionGRv1Alpha1())
	register[*mlApi.ArangoMLExtension](ArangoMLExtensionGKv1Beta1(), ArangoMLExtensionGRv1Beta1())
}

func ArangoMLExtensionGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoMLExtensionGroup,
		Kind:  ArangoMLExtensionKind,
	}
}

func ArangoMLExtensionGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLExtensionGroup,
		Kind:    ArangoMLExtensionKind,
		Version: ArangoMLExtensionVersionV1Alpha1,
	}
}

func ArangoMLExtensionGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLExtensionGroup,
		Kind:    ArangoMLExtensionKind,
		Version: ArangoMLExtensionVersionV1Beta1,
	}
}

func ArangoMLExtensionGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoMLExtensionGroup,
		Resource: ArangoMLExtensionResource,
	}
}

func ArangoMLExtensionGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLExtensionGroup,
		Resource: ArangoMLExtensionResource,
		Version:  ArangoMLExtensionVersionV1Alpha1,
	}
}

func ArangoMLExtensionGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLExtensionGroup,
		Resource: ArangoMLExtensionResource,
		Version:  ArangoMLExtensionVersionV1Beta1,
	}
}
