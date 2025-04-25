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

// ArangoMLStorage
const (
	ArangoMLStorageGroup           = ml.ArangoMLGroupName
	ArangoMLStorageResource        = ml.ArangoMLStorageResourcePlural
	ArangoMLStorageKind            = ml.ArangoMLStorageResourceKind
	ArangoMLStorageVersionV1Alpha1 = mlApiv1alpha1.ArangoMLVersion
	ArangoMLStorageVersionV1Beta1  = mlApi.ArangoMLVersion
)

func init() {
	register[*mlApiv1alpha1.ArangoMLStorage](ArangoMLStorageGKv1Alpha1(), ArangoMLStorageGRv1Alpha1())
	register[*mlApi.ArangoMLStorage](ArangoMLStorageGKv1Beta1(), ArangoMLStorageGRv1Beta1())
}

func ArangoMLStorageGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoMLStorageGroup,
		Kind:  ArangoMLStorageKind,
	}
}

func ArangoMLStorageGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLStorageGroup,
		Kind:    ArangoMLStorageKind,
		Version: ArangoMLStorageVersionV1Alpha1,
	}
}

func ArangoMLStorageGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLStorageGroup,
		Kind:    ArangoMLStorageKind,
		Version: ArangoMLStorageVersionV1Beta1,
	}
}

func ArangoMLStorageGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoMLStorageGroup,
		Resource: ArangoMLStorageResource,
	}
}

func ArangoMLStorageGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLStorageGroup,
		Resource: ArangoMLStorageResource,
		Version:  ArangoMLStorageVersionV1Alpha1,
	}
}

func ArangoMLStorageGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLStorageGroup,
		Resource: ArangoMLStorageResource,
		Version:  ArangoMLStorageVersionV1Beta1,
	}
}
