//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoBackupPolicySpec struct {
	// Schedule is cron-compatible specification of backup schedule
	Schedule string `json:"schedule"`
	// AllowConcurrent if false, ArangoBackup will not be created when previous Backups are not finished. Defaults to true
	AllowConcurrent *bool `json:"allowConcurrent,omitempty"`
	// DeploymentSelector specifies which deployments should get a backup
	DeploymentSelector *meta.LabelSelector `json:"selector,omitempty"`
	// ArangoBackupTemplate specifies additional options for newly created ArangoBackup
	BackupTemplate ArangoBackupTemplate `json:"template"`
}

// GetAllowConcurrent returns AllowConcurrent values. If AllowConcurrent is nil returns true
func (a ArangoBackupPolicySpec) GetAllowConcurrent() bool {
	return util.TypeOrDefault(a.AllowConcurrent, true)
}

type ArangoBackupTemplate struct {
	Options *ArangoBackupSpecOptions `json:"options,omitempty"`

	Upload *ArangoBackupSpecOperation `json:"upload,omitempty"`
}
