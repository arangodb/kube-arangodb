//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var logger = logging.Global().RegisterAndGetLogger("backup-operator", logging.Info)

const (
	// StateChange name of the event send when state changed
	StateChange = "StateChange"

	// FinalizerChange name of the event send when finalizer removed entry
	FinalizerChange = "FinalizerChange"
)

type handler struct {
	lock  sync.Mutex
	locks map[string]*sync.Mutex

	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	arangoClientFactory ArangoClientFactory

	operator operator.Operator
}

func (h *handler) Start(stopCh <-chan struct{}) {
	go h.start(stopCh)
}

func (h *handler) start(stopCh <-chan struct{}) {
	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-t.C:
			logger.Debug("Refreshing database objects")
			if err := h.refresh(); err != nil {
				log.Error().Err(err).Msgf("Unable to refresh database objects")
			}
			logger.Debug("Database objects refreshed")
		}
	}
}

func (h *handler) refresh() error {
	deployments, err := h.client.DatabaseV1().ArangoDeployments(h.operator.Namespace()).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		if err = h.refreshDeployment(&deployment); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) refreshDeployment(deployment *api.ArangoDeployment) error {
	m := h.getDeploymentMutex(deployment.Namespace, deployment.Name)
	m.Lock()
	defer m.Unlock()

	client, err := h.arangoClientFactory(deployment, nil)
	if err != nil {
		return err
	}

	backups, err := h.client.BackupV1().ArangoBackups(deployment.Namespace).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return err
	}

	for _, backup := range backups.Items {
		switch backup.GetStatus().ArangoBackupState.State {
		case backupApi.ArangoBackupStateCreate, backupApi.ArangoBackupStateCreating:
			// Skip refreshing backups if they are in creation state
			return nil
		}
	}

	existingBackups, err := client.List()
	if err != nil {
		return err
	}

	for _, backupMeta := range existingBackups {
		if err = h.refreshDeploymentBackup(deployment, backupMeta, backups.Items); err != nil {
			return err
		}
	}

	if err = h.cleanupImportedBackups(backups.Items); err != nil {
		return err
	}

	return nil
}

func (h *handler) cleanupImportedBackups(backups []backupApi.ArangoBackup) error {
	if !features.BackupCleanup().Enabled() {
		return nil
	}
	for _, backup := range backups {
		if backup.GetDeletionTimestamp() != nil {
			continue
		}

		if b := backup.Status.Backup; b == nil || !util.TypeOrDefault(b.Imported, false) {
			continue
		}

		logger.Str("name", backup.GetName()).Str("namespace", backup.GetNamespace()).Info("Removing Imported ArangoBackup")

		err := h.client.BackupV1().ArangoBackups(backup.GetNamespace()).Delete(context.Background(), backup.GetName(), meta.DeleteOptions{})
		if err != nil {
			return err
		}

	}

	return nil
}

func (h *handler) refreshDeploymentBackup(deployment *api.ArangoDeployment, backupMeta driver.BackupMeta, backups []backupApi.ArangoBackup) error {
	for _, backup := range backups {
		if download := backup.Spec.Download; download != nil {
			if download.ID == string(backupMeta.ID) {
				return nil
			}
		}

		if backup.Status.Backup == nil {
			continue
		}

		if backup.Status.Backup.ID == string(backupMeta.ID) {
			return nil
		}
	}

	name := fmt.Sprintf("backup-%s", uuid.NewUUID())

	logger.Str("id", string(backupMeta.ID)).Strs("namespace", deployment.GetNamespace()).Str("namespace", name).Info("Importing ArangoBackup from API")

	// New backup found, need to recreate
	backup := &backupApi.ArangoBackup{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: deployment.GetNamespace(),
		},
		Spec: backupApi.ArangoBackupSpec{
			Deployment: backupApi.ArangoBackupSpecDeployment{
				Name: deployment.Name,
			},
		},
	}

	_, err := h.client.BackupV1().ArangoBackups(backup.Namespace).Create(context.Background(), backup, meta.CreateOptions{})
	if err != nil {
		return err
	}

	status := updateStatus(backup,
		updateStatusState(backupApi.ArangoBackupStateReady, ""),
		updateStatusBackup(backupMeta),
		updateStatusBackupImported(util.NewType[bool](true)))

	_, err = operator.WithArangoBackupUpdateStatusInterfaceRetry(context.Background(), h.client.BackupV1().ArangoBackups(backup.GetNamespace()), backup, *status, meta.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Name() string {
	return backup.ArangoBackupResourceKind
}

func (h *handler) getDeploymentMutex(namespace, deployment string) *sync.Mutex {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.locks == nil {
		h.locks = map[string]*sync.Mutex{}
	}

	name := fmt.Sprintf("%s/%s", namespace, deployment)

	if _, ok := h.locks[name]; !ok {
		h.locks[name] = &sync.Mutex{}
	}

	return h.locks[name]
}

func (h *handler) Handle(_ context.Context, item operation.Item) error {
	// Get object. It also covers NotFound case
	b, err := h.client.BackupV1().ArangoBackups(item.Namespace).Get(context.Background(), item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	// Check if we should start finalizer
	if b.DeletionTimestamp != nil {
		logger.Debug("Finalizing %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return h.finalize(b)
	}

	// Add finalizers
	if !hasFinalizers(b) {
		b.Finalizers = appendFinalizers(b)
		log.Info().Msgf("Updating finalizers %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		if _, err = h.client.BackupV1().ArangoBackups(item.Namespace).Update(context.Background(), b, meta.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	}

	// Create lock per namespace to ensure that we are not using two goroutines in same time
	lock := h.getDeploymentMutex(b.Namespace, b.Spec.Deployment.Name)
	lock.Lock()
	defer lock.Unlock()

	// Add owner reference
	if len(b.OwnerReferences) == 0 {
		deployment, err := h.client.DatabaseV1().ArangoDeployments(b.Namespace).Get(context.Background(), b.Spec.Deployment.Name, meta.GetOptions{})
		if err == nil {
			b.OwnerReferences = []meta.OwnerReference{
				deployment.AsOwner(),
			}

			if _, err = h.client.BackupV1().ArangoBackups(item.Namespace).Update(context.Background(), b, meta.UpdateOptions{}); err != nil {
				return err
			}
		}

		b, err = h.client.BackupV1().ArangoBackups(item.Namespace).Get(context.Background(), item.Name, meta.GetOptions{})
		if err != nil {
			if apiErrors.IsNotFound(err) {
				return nil
			}

			return err
		}
	}

	status, err := h.processArangoBackup(b.DeepCopy())
	if err != nil {
		log.Warn().Err(err).Msgf("Fail for %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		cError := switchError(err)

		var temporaryError temporaryError
		if errors.As(cError, &temporaryError) {
			return cError
		}

		status, _ = setFailedState(b, cError)
	}

	if status == nil {
		return nil
	}

	// Nothing to update, objects are equal
	if b.Status.Equal(status) {
		return nil
	}

	if h.operator != nil {
		h.operator.EnqueueItem(item)
	}

	// Ensure that transit is possible
	if err = backupApi.ArangoBackupStateMap.Transit(b.Status.State, status.State); err != nil {
		return err
	}

	// Log message about state change
	if b.Status.State != status.State {
		if status.State == backupApi.ArangoBackupStateFailed {
			h.eventRecorder.Warning(b, StateChange, "Transiting from %s to %s with error: %s",
				b.Status.State,
				status.State,
				status.Message)
		} else {
			if status.Message != "" {
				h.eventRecorder.Normal(b, StateChange, "Transiting from %s to %s with message: %s",
					b.Status.State,
					status.State,
					status.Message)
			} else {
				h.eventRecorder.Normal(b, StateChange, "Transiting from %s to %s",
					b.Status.State,
					status.State)
			}
		}
	}

	logger.Debug("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	// Update status on object
	if _, err := operator.WithArangoBackupUpdateStatusInterfaceRetry(context.Background(), h.client.BackupV1().ArangoBackups(b.GetNamespace()), b, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (h *handler) processArangoBackup(backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	if err := backup.Validate(); err != nil {
		return setFailedState(backup, err)
	}

	if f, ok := stateHolders[backup.Status.State]; ok {
		return f(h, backup)
	}

	return nil, errors.Errorf("state %s is not supported", backup.Status.State)
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == backupApi.SchemeGroupVersion.Group &&
		item.Version == backupApi.SchemeGroupVersion.Version &&
		item.Kind == backup.ArangoBackupResourceKind
}

func (h *handler) getArangoDeploymentObject(backup *backupApi.ArangoBackup) (*api.ArangoDeployment, error) {
	if backup.Spec.Deployment.Name == "" {
		return nil, newFatalErrorf("deployment ref is not specified for backup %s/%s", backup.Namespace, backup.Name)
	}

	obj, err := h.client.DatabaseV1().ArangoDeployments(backup.Namespace).Get(context.Background(), backup.Spec.Deployment.Name, meta.GetOptions{})
	if err == nil {
		return obj, nil
	}

	// Check if object is not found
	if apiErrors.IsNotFound(err) {
		return nil, newFatalError(err)
	}

	// Otherwise it is connection issue - mark as temporary
	return nil, newTemporaryError(err)
}
