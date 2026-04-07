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

package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type arangoClientBackupImpl struct {
	deployment *api.ArangoDeployment
	backup     *backupApi.ArangoBackup
	driver     adbDriverV2.Client
	kubecli    kubernetes.Interface
}

func newArangoClientBackupFactory(handler *handler) ArangoClientFactory {
	return func(deployment *api.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		ctx := context.Background()
		client, err := arangod.CreateArangodDatabaseClient(ctx, handler.kubeClient.CoreV1(), deployment, false, true)
		if err != nil {
			return nil, err
		}

		return &arangoClientBackupImpl{
			deployment: deployment,
			backup:     backup,
			driver:     client,
			kubecli:    handler.kubeClient,
		}, nil
	}
}

func (ac *arangoClientBackupImpl) List() (map[string]adbDriverV2.BackupMeta, error) {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	backups, err := ac.driver.BackupList(ctx, nil)
	if err != nil {
		return nil, err
	}

	return backups.Backups, nil
}

func (ac *arangoClientBackupImpl) Create() (ArangoBackupCreateResponse, error) {
	dt := globals.GetGlobalTimeouts().BackupArangoClientTimeout().Get()

	co := adbDriverV2.BackupCreateOptions{}

	if opt := ac.backup.Spec.Options; opt != nil {
		if allowInconsistent := opt.AllowInconsistent; allowInconsistent != nil {
			co.AllowInconsistent = allowInconsistent
		}
		if timeout := opt.Timeout; timeout != nil {
			co.Timeout = util.NewType(uint(*timeout))
			dt += time.Duration(*timeout * float32(time.Second))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()

	resp, err := ac.driver.BackupCreate(ctx, &co)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	// Now ask for the version
	meta, err := ac.Get(resp.ID)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	return ArangoBackupCreateResponse{
		PotentiallyInconsistent: resp.PotentiallyInconsistent,
		BackupMeta:              meta,
	}, nil
}

func (ac *arangoClientBackupImpl) CreateAsync(jobID string) (ArangoBackupCreateResponse, error) {
	dt := globals.GetGlobalTimeouts().BackupArangoClientTimeout().Get()

	co := adbDriverV2.BackupCreateOptions{}

	if opt := ac.backup.Spec.Options; opt != nil {
		if allowInconsistent := opt.AllowInconsistent; allowInconsistent != nil {
			co.AllowInconsistent = allowInconsistent
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()

	if jobID == "" {
		ctx = adbDriverV2Connection.WithAsync(ctx)
	} else {
		ctx = adbDriverV2Connection.WithAsyncID(ctx, jobID)
	}

	resp, err := ac.driver.BackupCreate(ctx, &co)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	// Now ask for the version
	meta, err := ac.Get(resp.ID)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	return ArangoBackupCreateResponse{
		PotentiallyInconsistent: resp.PotentiallyInconsistent,
		BackupMeta:              meta,
	}, nil
}

func (ac *arangoClientBackupImpl) Get(backupID string) (adbDriverV2.BackupMeta, error) {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	// list, err := ac.driver.Backup().List(ctx, &driver.BackupListOptions{ID: backupID})
	list, err := ac.driver.BackupList(ctx, nil)
	if err != nil {
		return adbDriverV2.BackupMeta{}, err
	}

	meta, ok := list.Backups[backupID]

	if ok {
		return meta, nil
	}

	return adbDriverV2.BackupMeta{}, adbDriverV2Shared.ArangoError{
		ErrorMessage: fmt.Sprintf("backup %s was not found", backupID),
		Code:         404,
	}
}

func (ac *arangoClientBackupImpl) getCredentialsFromSecret(ctx context.Context, secretName string) (interface{}, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	token, err := k8sutil.GetTokenSecretString(ctxChild, ac.kubecli.CoreV1().Secrets(ac.backup.Namespace), secretName)
	if err != nil {
		return nil, err
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(token), &raw); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal credentials: ")
	}

	return raw, nil
}

func (ac *arangoClientBackupImpl) Upload(backupID string) (string, error) {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientUploadTimeout().WithTimeout(context.Background())
	defer cancel()

	uploadSpec := ac.backup.Spec.Upload
	if uploadSpec == nil {
		return "", errors.Errorf("upload was called but no upload spec was given")
	}

	cred, err := ac.getCredentialsFromSecret(ctx, uploadSpec.CredentialsSecretName)
	if err != nil {
		return "", err
	}

	if transfer, err := ac.driver.BackupUpload(ctx, backupID, uploadSpec.RepositoryURL, cred); err != nil {
		return "", err
	} else {
		return transfer.GetID(), err
	}
}

func (ac *arangoClientBackupImpl) Download(backupID string) (string, error) {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientUploadTimeout().WithTimeout(context.Background())
	defer cancel()

	downloadSpec := ac.backup.Spec.Download
	if downloadSpec == nil {
		return "", errors.Errorf("Download was called but not download spec was given")
	}

	cred, err := ac.getCredentialsFromSecret(ctx, downloadSpec.CredentialsSecretName)
	if err != nil {
		return "", err
	}

	if transfer, err := ac.driver.BackupDownload(ctx, backupID, downloadSpec.RepositoryURL, cred); err != nil {
		return "", err
	} else {
		return transfer.GetID(), err
	}
}

func (ac *arangoClientBackupImpl) Progress(jobID string, transferType adbDriverV2.TransferType) (ArangoBackupProgress, error) {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	mon, err := ac.driver.TransferMonitor(jobID, transferType)
	if err != nil {
		return ArangoBackupProgress{}, err
	}

	report, err := mon.Progress(ctx)
	if err != nil {
		return ArangoBackupProgress{}, err
	}

	if report.Cancelled {
		return ArangoBackupProgress{
			Failed:      true,
			FailMessage: "Upload cancelled",
		}, nil
	}

	var ret ArangoBackupProgress
	var completedCount int
	var total int
	var done int

	for _, status := range report.DBServers {
		total += status.Progress.Total
		done += status.Progress.Done

		switch status.Status {
		case adbDriverV2.TransferFailed:
			ret.Failed = true
			ret.FailMessage = status.ErrorMessage
		case adbDriverV2.TransferCompleted:
			completedCount++
		case adbDriverV2.TransferAcknowledged:
		case adbDriverV2.TransferStarted:
		case "":
			completedCount++
		default:
			return ArangoBackupProgress{}, errors.Errorf("Unknown transfer status: %s", status.Status)
		}
	}

	// Check if all defined servers are completed and the total number of files is greater than 0 (there is at least 1 file per server)
	ret.Completed = completedCount == len(report.DBServers) && total > 0
	if total != 0 {
		ret.Progress = (100 * done) / total
	}
	return ret, nil
}

func (ac *arangoClientBackupImpl) Exists(backupID string) (bool, error) {
	_, err := ac.Get(backupID)
	if err != nil {
		if adbDriverV2Shared.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (ac *arangoClientBackupImpl) Delete(backupID string) error {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	return ac.driver.BackupDelete(ctx, backupID)
}

func (ac *arangoClientBackupImpl) Abort(jobID string, transferType adbDriverV2.TransferType) error {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	mon, err := ac.driver.TransferMonitor(jobID, transferType)
	if err != nil {
		return err
	}

	return mon.Abort(ctx)
}

func (ac *arangoClientBackupImpl) HealthCheck() error {
	ctx, cancel := globals.GetGlobalTimeouts().BackupArangoClientTimeout().WithTimeout(context.Background())
	defer cancel()

	_, err := ac.driver.Version(ctx)
	return err
}
