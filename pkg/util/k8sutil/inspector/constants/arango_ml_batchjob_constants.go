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
)

// ArangoMLBatchJob
const (
	ArangoMLBatchJobGroup           = ml.ArangoMLGroupName
	ArangoMLBatchJobResource        = ml.ArangoMLBatchJobResourcePlural
	ArangoMLBatchJobKind            = ml.ArangoMLBatchJobResourceKind
	ArangoMLBatchJobVersionV1Alpha1 = mlApiv1alpha1.ArangoMLVersion
)

func init() {
	register[*mlApiv1alpha1.ArangoMLBatchJob](ArangoMLBatchJobGKv1Alpha1(), ArangoMLBatchJobGRv1Alpha1())
}

func ArangoMLBatchJobGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoMLBatchJobGroup,
		Kind:  ArangoMLBatchJobKind,
	}
}

func ArangoMLBatchJobGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLBatchJobGroup,
		Kind:    ArangoMLBatchJobKind,
		Version: ArangoMLBatchJobVersionV1Alpha1,
	}
}

func ArangoMLBatchJobGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoMLBatchJobGroup,
		Resource: ArangoMLBatchJobResource,
	}
}

func ArangoMLBatchJobGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLBatchJobGroup,
		Resource: ArangoMLBatchJobResource,
		Version:  ArangoMLBatchJobVersionV1Alpha1,
	}
}
