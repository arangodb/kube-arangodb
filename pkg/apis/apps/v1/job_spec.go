//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package v1

import batch "k8s.io/api/batch/v1"

type ArangoJobSpec struct {
	// ArangoDeploymentName holds the name of ArangoDeployment
	ArangoDeploymentName string `json:"arangoDeploymentName"`

	// JobTemplate holds the Kubernetes Job Template
	// +doc/type: batch.JobSpec
	// +doc/link: Kubernetes Documentation|https://kubernetes.io/docs/concepts/workloads/controllers/job/
	// +doc/link: Documentation of batch.JobSpec|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#jobspec-v1-batch
	JobTemplate *batch.JobSpec `json:"jobTemplate,omitempty"`
}
