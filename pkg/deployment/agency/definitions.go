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

package agency

import (
	"fmt"
	"strings"
)

type ReadRequest [][]string

const (
	ArangoKey   = "arango"
	ArangoDBKey = "arangodb"

	PlanKey    = "Plan"
	CurrentKey = "Current"
	TargetKey  = "Target"

	CurrentMaintenanceServers = "MaintenanceServers"

	TargetHotBackupKey = "HotBackup"

	PlanCollectionsKey = "Collections"
	PlanDatabasesKey   = "Databases"

	SupervisionKey            = "Supervision"
	SupervisionMaintenanceKey = "Maintenance"

	TargetJobToDoKey     = "ToDo"
	TargetJobPendingKey  = "Pending"
	TargetJobFailedKey   = "Failed"
	TargetJobFinishedKey = "Finished"

	TargetCleanedServersKey = "CleanedServers"

	ArangoSyncKey                     = "arangosync"
	ArangoSyncStateKey                = "synchronizationState"
	ArangoSyncStateIncomingKey        = "incoming"
	ArangoSyncStateIncomingStateKey   = "state"
	ArangoSyncStateOutgoingKey        = "outgoing"
	ArangoSyncStateOutgoingTargetsKey = "targets"
)

func GetAgencyKey(parts ...string) string {
	return fmt.Sprintf("/%s", strings.Join(parts, "/"))
}

func GetAgencyReadKey(elements ...string) []string {
	return elements
}

func GetAgencyReadRequest(elements ...[]string) ReadRequest {
	return elements
}

func GetAgencyReadRequestFields() ReadRequest {
	return GetAgencyReadRequest([]string{
		GetAgencyKey(ArangoKey, SupervisionKey, SupervisionMaintenanceKey),
		GetAgencyKey(ArangoKey, PlanKey, PlanCollectionsKey),
		GetAgencyKey(ArangoKey, PlanKey, PlanDatabasesKey),
		GetAgencyKey(ArangoKey, CurrentKey, PlanCollectionsKey),
		GetAgencyKey(ArangoKey, CurrentKey, CurrentMaintenanceServers),
		GetAgencyKey(ArangoKey, TargetKey, TargetHotBackupKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobToDoKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobPendingKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobFailedKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobFinishedKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetCleanedServersKey),
		GetAgencyKey(ArangoDBKey, ArangoSyncKey, ArangoSyncStateKey, ArangoSyncStateIncomingKey, ArangoSyncStateIncomingStateKey),
		GetAgencyKey(ArangoDBKey, ArangoSyncKey, ArangoSyncStateKey, ArangoSyncStateOutgoingKey, ArangoSyncStateOutgoingTargetsKey),
	})
}
