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

package v2alpha1

type DeploymentMemberPropagationMode string

func (d *DeploymentMemberPropagationMode) Get() DeploymentMemberPropagationMode {
	if d == nil {
		return DeploymentMemberPropagationModeDefault
	}

	return *d
}

func (d DeploymentMemberPropagationMode) New() *DeploymentMemberPropagationMode {
	return &d
}

func (d DeploymentMemberPropagationMode) String() string {
	return string(d)
}

func (d *DeploymentMemberPropagationMode) Equal(b *DeploymentMemberPropagationMode) bool {
	if d == nil && b == nil {
		return true
	}

	if d == nil || b == nil {
		return false
	}

	return *d == *b
}

const (
	// DeploymentMemberPropagationModeDefault Define default propagation mode
	DeploymentMemberPropagationModeDefault = DeploymentMemberPropagationModeAlways
	// DeploymentMemberPropagationModeAlways define mode which restart member whenever change in pod is discovered
	DeploymentMemberPropagationModeAlways DeploymentMemberPropagationMode = "always"
	// DeploymentMemberPropagationModeOnRestart propagate member spec whenever pod is restarted. Do not restart member by default
	DeploymentMemberPropagationModeOnRestart DeploymentMemberPropagationMode = "on-restart"
)
