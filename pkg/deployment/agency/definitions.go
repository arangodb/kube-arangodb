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

package agency

import (
	"fmt"
	"strings"
)

const (
	ArangoKey = "arango"

	PlanKey    = "Plan"
	CurrentKey = "Current"
	TargetKey  = "Target"

	CurrentMaintenanceServers = "MaintenanceServers"

	TargetHotBackupKey = "HotBackup"

	PlanCollectionsKey = "Collections"

	SupervisionKey            = "Supervision"
	SupervisionMaintenanceKey = "Maintenance"

	TargetJobToDoKey     = "ToDo"
	TargetJobPendingKey  = "Pending"
	TargetJobFailedKey   = "Failed"
	TargetJobFinishedKey = "Finished"

	TargetCleanedServersKey = "CleanedServers"
)

func GetAgencyKey(parts ...string) string {
	return fmt.Sprintf("/%s", strings.Join(parts, "/"))
}

func GetAgencyReadKey(elements ...string) []string {
	return elements
}

func GetAgencyReadRequest(elements ...[]string) [][]string {
	return elements
}
