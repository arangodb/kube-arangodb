//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

// ConditionType is a strongly typed condition name
type ConditionType = sharedApi.ConditionType

const (
	// ConditionTypeConfigured indicates that the replication has been configured.
	ConditionTypeConfigured ConditionType = "Configured"
	// ConditionTypeEnsuredInSync indicates that the replication consistency was checked.
	ConditionTypeEnsuredInSync ConditionType = "EnsuredInSync"
	// ConditionTypeAborted indicates that the replication was canceled with abort option.
	ConditionTypeAborted ConditionType = "Aborted"

	// ConditionConfiguredReasonActive describes synchronization as active.
	ConditionConfiguredReasonActive = "Active"
	// ConditionConfiguredReasonInactive describes synchronization as inactive.
	ConditionConfiguredReasonInactive = "Inactive"
	// ConditionConfiguredReasonInvalid describes synchronization as active.
	ConditionConfiguredReasonInvalid = "Invalid"
)

// Condition represents one current condition of a deployment or deployment member.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
type Condition = sharedApi.Condition

// ConditionList is a list of conditions.
// Each type is allowed only once.
type ConditionList = sharedApi.ConditionList
