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

// ArangoMLCronJob
const (
	ArangoMLCronJobGroup           = ml.ArangoMLGroupName
	ArangoMLCronJobResource        = ml.ArangoMLCronJobResourcePlural
	ArangoMLCronJobKind            = ml.ArangoMLCronJobResourceKind
	ArangoMLCronJobVersionV1Alpha1 = mlApiv1alpha1.ArangoMLVersion
)

func init() {
	register[*mlApiv1alpha1.ArangoMLCronJob](ArangoMLCronJobGKv1Alpha1(), ArangoMLCronJobGRv1Alpha1())
}

func ArangoMLCronJobGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoMLCronJobGroup,
		Kind:  ArangoMLCronJobKind,
	}
}

func ArangoMLCronJobGKv1Alpha1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoMLCronJobGroup,
		Kind:    ArangoMLCronJobKind,
		Version: ArangoMLCronJobVersionV1Alpha1,
	}
}

func ArangoMLCronJobGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoMLCronJobGroup,
		Resource: ArangoMLCronJobResource,
	}
}

func ArangoMLCronJobGRv1Alpha1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoMLCronJobGroup,
		Resource: ArangoMLCronJobResource,
		Version:  ArangoMLCronJobVersionV1Alpha1,
	}
}
