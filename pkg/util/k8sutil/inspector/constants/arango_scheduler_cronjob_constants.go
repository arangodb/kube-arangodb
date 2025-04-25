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

// ArangoSchedulerCronJob
const (
	ArangoSchedulerCronJobGroup          = scheduler.ArangoSchedulerGroupName
	ArangoSchedulerCronJobResource       = scheduler.CronJobResourcePlural
	ArangoSchedulerCronJobKind           = scheduler.CronJobResourceKind
	ArangoSchedulerCronJobVersionV1Beta1 = schedulerApi.ArangoSchedulerVersion
)

func init() {
	register[*schedulerApi.ArangoSchedulerCronJob](ArangoSchedulerCronJobGKv1Beta1(), ArangoSchedulerCronJobGRv1Beta1())
}

func ArangoSchedulerCronJobGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoSchedulerCronJobGroup,
		Kind:  ArangoSchedulerCronJobKind,
	}
}

func ArangoSchedulerCronJobGKv1Beta1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoSchedulerCronJobGroup,
		Kind:    ArangoSchedulerCronJobKind,
		Version: ArangoSchedulerCronJobVersionV1Beta1,
	}
}

func ArangoSchedulerCronJobGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoSchedulerCronJobGroup,
		Resource: ArangoSchedulerCronJobResource,
	}
}

func ArangoSchedulerCronJobGRv1Beta1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoSchedulerCronJobGroup,
		Resource: ArangoSchedulerCronJobResource,
		Version:  ArangoSchedulerCronJobVersionV1Beta1,
	}
}
