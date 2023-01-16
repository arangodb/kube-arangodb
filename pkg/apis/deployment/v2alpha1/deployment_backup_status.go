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

package v2alpha1

type DeploymentRestoreState string

const (
	DeploymentRestoreStateRestoring     DeploymentRestoreState = "Restoring"
	DeploymentRestoreStateRestored      DeploymentRestoreState = "Restored"
	DeploymentRestoreStateRestoreFailed DeploymentRestoreState = "RestoreFailed"
)

type DeploymentRestoreResult struct {
	RequestedFrom string                 `json:"requestedFrom"`
	State         DeploymentRestoreState `json:"state"`
	Message       string                 `json:"message,omitempty"`
}

func (dr *DeploymentRestoreResult) Equal(other *DeploymentRestoreResult) bool {
	if dr == nil {
		return other == nil
	}

	if other == nil {
		return false
	}

	return dr.RequestedFrom == other.RequestedFrom &&
		dr.Message == other.Message &&
		dr.State == other.State
}
