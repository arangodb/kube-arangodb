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
	"reflect"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/backup/event"

	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rs/zerolog/log"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultArangoClientTimeout = 30 * time.Second

	StateChange     = "StateChange"
	FinalizerChange = "FinalizerChange"
)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.EventRecorderInstance

	arangoClientFactory ArangoClientFactory
	arangoClientTimeout time.Duration
}

func (h *handler) Name() string {
	return database.ArangoBackupResourceKind
}

func (h *handler) Handle(item operator.Item) error {
	// Get Backup object. It also cover NotFound case
	backup, err := h.client.DatabaseV1alpha().ArangoBackups(item.Namespace).Get(item.Name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	// Check if we should start finalizer
	if backup.DeletionTimestamp != nil {
		log.Debug().Msgf("Finalizing %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return h.finalize(backup)
	}

	// Do not act on delete event, finalizer should be used
	if item.Operation == operator.OperationDelete {
		return nil
	}

	// Add finalizers
	if !hasFinalizers(backup) {
		backup.Finalizers = appendFinalizers(backup)
		log.Info().Msgf("Updating finalizers %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		if _, err = h.client.DatabaseV1alpha().ArangoBackups(item.Namespace).Update(backup); err != nil {
			return err
		}

		return nil
	}

	status, err := h.processArangoBackup(backup.DeepCopy())
	if err != nil {
		return err
	}

	// Nothing to update, objects are equal
	if reflect.DeepEqual(backup.Status, status) {
		return nil
	}

	// Ensure that transit is possible
	if err = database.ArangoBackupStateMap.Transit(backup.Status.State, status.State); err != nil {
		return err
	}

	// Log message about state change
	if backup.Status.State != status.State {
		if status.State == database.ArangoBackupStateFailed {
			h.eventRecorder.Warning(backup, StateChange, "Transiting from %s to %s with error: %s",
				backup.Status.State,
				status.State,
				status.Message)
		} else {
			h.eventRecorder.Normal(backup, StateChange, "Transiting from %s to %s",
				backup.Status.State,
				status.State)
		}
	} else {
		// Keep old time in case when object did not change
		status.Time = backup.Status.Time
	}

	backup.Status = status

	log.Debug().Msgf("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	// Update status on object
	if _, err = h.client.DatabaseV1alpha().ArangoBackups(item.Namespace).UpdateStatus(backup); err != nil {
		return err
	}

	return nil
}

func (h *handler) processArangoBackup(backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	if err := backup.Validate(); err != nil {
		return createFailedState(err, backup.Status), nil
	}

	if f, ok := stateHolders[backup.Status.State]; !ok {
		return database.ArangoBackupStatus{}, fmt.Errorf("state %s is not supported", backup.Status.State)
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
