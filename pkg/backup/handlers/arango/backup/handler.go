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
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	defaultArangoClientTimeout = 30*time.Second
)

type handler struct {
	client arangoClientSet.Interface

	arangoClientFactory ArangoClientFactory
	arangoClientTimeout time.Duration
}

func (h *handler) Name() string {
	return "ArangoBackup"
}

func (h *handler) Handle(item operator.Item) error {
	// Do not act on delete event, finalizers are used
	if item.Operation == operator.OperationDelete {
		return nil
	}

	// Get Backup object. It also cover NotFound case
	backup, err := h.client.DatabaseV1alpha().ArangoBackups(item.Namespace).Get(item.Name, meta.GetOptions{})
	if err != nil {
		return err
	}

	// Check if object is valid. With AdmissionWebhooks this state should always return true
	if err = backup.Validate(); err != nil {
		return err
	}

	status, err := h.processArangoBackup(backup.DeepCopy())
	if err != nil {
		return err
	}

	// Ensure that transit is possible
	if err = database.ArangoBackupStateMap.Transit(backup.Status.State.State, status.State.State); err != nil {
		return err
	}

	backup.Status = status

	// Update status on object
	if _, err = h.client.DatabaseV1alpha().ArangoBackups(item.Namespace).UpdateStatus(backup); err != nil {
		return err
	}

	return nil
}

func (h *handler) processArangoBackup(backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	if f, ok := stateHolders[backup.Status.State.State]; !ok {
		return database.ArangoBackupStatus{}, fmt.Errorf("state %s is not supported", backup.Status.State.State)
	} else {
		return f(h, backup)
	}
}

func (h *handler) CanBeHandled(item operator.Item) bool {
	return item.Group == database.SchemeGroupVersion.Group &&
		item.Version == database.SchemeGroupVersion.Version &&
		item.Kind == database.ArangoBackupResourceKind
}

func (h *handler) getArangoDeploymentObject(backup *database.ArangoBackup) (*database.ArangoDeployment, error) {
	if backup.Spec.Deployment.Name == "" {
		return nil, fmt.Errorf("deployment ref is not specified for backup %s/%s", backup.Namespace, backup.Name)
	}

	return h.client.DatabaseV1alpha().ArangoDeployments(backup.Namespace).Get(backup.Spec.Deployment.Name, meta.GetOptions{})
}