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
	// Parsed by https://godoc.org/github.com/robfig/cron
	Schedule string `json:"schedule"`
	// AllowConcurrent if false, ArangoBackup will not be created when previous Backups are not finished
	// +doc/default: true
	AllowConcurrent *bool `json:"allowConcurrent,omitempty"`
	// DeploymentSelector Selector definition for selecting matching ArangoBackup Custom Resources.
	// +doc/type: meta.LabelSelector
	// +doc/link: Kubernetes Documentation|https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector
	DeploymentSelector *meta.LabelSelector `json:"selector,omitempty"`
	// MaxBackups defines how many backups should be kept in history (per deployment). Oldest healthy Backups will be deleted.
	// If not specified or 0 then no limit is applied
	// +doc/default: 0
	MaxBackups int `json:"maxBackups,omitempty"`
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

	Backoff *ArangoBackupSpecBackOff `json:"backoff,omitempty"`

	// Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".
	Lifetime *meta.Duration `json:"lifetime,omitempty"`
}
