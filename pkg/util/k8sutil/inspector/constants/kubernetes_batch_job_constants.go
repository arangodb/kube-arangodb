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

// Job
const (
	JobGroup     = batch.GroupName
	JobResource  = "jobs"
	JobKind      = "Job"
	JobVersionV1 = "v1"
)

func init() {
	register[*batch.Job](JobGKv1(), JobGRv1())
}

func JobGK() schema.GroupKind {
	return schema.GroupKind{
		Group: JobGroup,
		Kind:  JobKind,
	}
}

func JobGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   JobGroup,
		Kind:    JobKind,
		Version: JobVersionV1,
	}
}

func JobGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    JobGroup,
		Resource: JobResource,
	}
}

func JobGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    JobGroup,
		Resource: JobResource,
		Version:  JobVersionV1,
	}
}
