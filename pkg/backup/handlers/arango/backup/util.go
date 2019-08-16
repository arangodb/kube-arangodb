//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package backup

import (
	"fmt"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
)

func switchTemporaryError(err error, status database.ArangoBackupStatus) (database.ArangoBackupStatus, error) {
	if IsTemporaryError(err) {
		return database.ArangoBackupStatus{}, err
	}

	return createFailedState(err, status), nil
}

func createFailMessage(state state.State, message string) string {
	return fmt.Sprintf("Failed State %s: %s", state, message)
}

func createFailedState(err error, status database.ArangoBackupStatus) database.ArangoBackupStatus {
	newStatus := status.DeepCopy()

	newStatus.ArangoBackupState = database.ArangoBackupState{
		State:   database.ArangoBackupStateFailed,
		Message: createFailMessage(status.State, err.Error()),
	}

	newStatus.Available = false

	return *newStatus
}
