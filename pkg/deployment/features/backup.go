//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

package features

func init() {
	registerFeature(asyncBackupCreation)
	registerFeature(backupCleanup)
	registerFeature(backupPolicyUntilPropagation)
}

var asyncBackupCreation = &feature{
	name:               "async-backup-creation",
	description:        "Create backups asynchronously to avoid blocking the operator and reaching the timeout",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

var backupCleanup = &feature{
	name:               "backup-cleanup",
	description:        "Cleanup imported backups if required",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

var backupPolicyUntilPropagation = &feature{
	name:               "backup-policy-until-propagation",
	description:        "Sets Until field in the Backup based on next schedule time",
	enterpriseRequired: false,
	enabledByDefault:   true,
}

// AsyncBackupCreation returns mode for backup creation (sync/async).
func AsyncBackupCreation() Feature {
	return asyncBackupCreation
}

// BackupCleanup returns mode for Imported backups cleanup.
func BackupCleanup() Feature {
	return backupCleanup
}

// BackupPolicyUntilPropagation returns mode for backup propagation.
func BackupPolicyUntilPropagation() Feature {
	return backupPolicyUntilPropagation
}
