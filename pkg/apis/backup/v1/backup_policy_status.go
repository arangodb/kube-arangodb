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
)

// ArangoBackupPolicyStatus Status of the ArangoBackupPolicy Custom Resource managed by operator
type ArangoBackupPolicyStatus struct {
	// Scheduled Next scheduled time in UTC
	// +doc/type: meta.Time
	Scheduled meta.Time `json:"scheduled,omitempty"`
	// Message from the operator in case of failures - schedule not valid, ArangoBackupPolicy not valid
	Message string `json:"message,omitempty"`
}
