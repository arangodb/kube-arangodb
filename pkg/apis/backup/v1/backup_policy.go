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

package v1

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	deployment "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoBackupPolicyList is a list of ArangoDB backup policy.
type ArangoBackupPolicyList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoBackupPolicy `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoBackupPolicy contains definition and status of the ArangoDB Backup Policy.
type ArangoBackupPolicy struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoBackupPolicySpec   `json:"spec"`
	Status ArangoBackupPolicyStatus `json:"status"`
}

func (a *ArangoBackupPolicy) NewBackup(d *deployment.ArangoDeployment) *ArangoBackup {
	policyName := a.Name

	spec := &ArangoBackupSpec{
		Deployment: ArangoBackupSpecDeployment{
			Name: d.Name,
		},
		Upload:     a.Spec.BackupTemplate.Upload.DeepCopy(),
		Options:    a.Spec.BackupTemplate.Options.DeepCopy(),
		PolicyName: &policyName,
	}

	return &ArangoBackup{
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", d.Name, utils.RandomString(8)),
			Namespace: a.Namespace,

			Labels:      d.Labels,
			Annotations: d.Annotations,

			Finalizers: []string{
				FinalizerArangoBackup,
			},
		},
		Spec: *spec,
	}
}
