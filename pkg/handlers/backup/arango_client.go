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

package backup

import (
	"net/http"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
)

var (
	temporaryErrorNum = utils.IntList{
		1465, // Communication error with server
	}

	temporaryCodes = utils.IntList{
		http.StatusServiceUnavailable,
	}
)

// ArangoClientFactory factory type for creating clients
type ArangoClientFactory func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error)

// ArangoBackupProgress progress info
type ArangoBackupProgress struct {
	Progress          int
	Failed, Completed bool
	FailMessage       string
}

// ArangoBackupCreateResponse create response
type ArangoBackupCreateResponse struct {
	driver.BackupMeta
	PotentiallyInconsistent bool
}

// ArangoBackupClient interface with backup functionality for database
type ArangoBackupClient interface {
	Create() (ArangoBackupCreateResponse, error)
	Get(driver.BackupID) (driver.BackupMeta, error)

	Upload(driver.BackupID) (driver.BackupTransferJobID, error)
	Download(driver.BackupID) (driver.BackupTransferJobID, error)

	Progress(driver.BackupTransferJobID) (ArangoBackupProgress, error)
	Abort(driver.BackupTransferJobID) error

	Exists(driver.BackupID) (bool, error)
	Delete(driver.BackupID) error

	List() (map[driver.BackupID]driver.BackupMeta, error)
}
