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
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CronJob
const (
	CronJobGroup     = batch.GroupName
	CronJobResource  = "cronjobs"
	CronJobKind      = "CronJob"
	CronJobVersionV1 = "v1"
)

func init() {
	register[*batch.CronJob](CronJobGKv1(), CronJobGRv1())
}

func CronJobGK() schema.GroupKind {
	return schema.GroupKind{
		Group: CronJobGroup,
		Kind:  CronJobKind,
	}
}

func CronJobGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   CronJobGroup,
		Kind:    CronJobKind,
		Version: CronJobVersionV1,
	}
}

func CronJobGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    CronJobGroup,
		Resource: CronJobResource,
	}
}

func CronJobGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    CronJobGroup,
		Resource: CronJobResource,
		Version:  CronJobVersionV1,
	}
}
