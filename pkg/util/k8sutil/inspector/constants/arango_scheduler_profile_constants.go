//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

// ArangoProfile
const (
	ArangoProfileGroup          = scheduler.ArangoSchedulerGroupName
	ArangoProfileResource       = scheduler.ArangoProfileResourcePlural
	ArangoProfileKind           = scheduler.ArangoProfileResourceKind
	ArangoProfileVersionV1Beta1 = schedulerApi.ArangoSchedulerVersion
)

func init() {
	register[*schedulerApi.ArangoProfile](ArangoProfileGKv1Beta1(), ArangoProfileGRv1Beta1())
}

func ArangoProfileGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoProfileGroup,
		Kind:  ArangoProfileKind,
	}
}

func ArangoProfileGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoProfileGroup,
		Kind:    ArangoProfileKind,
		Version: ArangoProfileVersionV1Beta1,
	}
}

func ArangoProfileGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoProfileGroup,
		Resource: ArangoProfileResource,
	}
}

func ArangoProfileGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoProfileGroup,
		Resource: ArangoProfileResource,
		Version:  ArangoProfileVersionV1Beta1,
	}
}
