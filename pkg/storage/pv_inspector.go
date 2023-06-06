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

package storage

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// inspectPVs queries all PersistentVolume's and triggers a cleanup for
// released volumes.
// Returns the number of available PV's.
func (ls *LocalStorage) inspectPVs() (int, error) {
	var volumes []*core.PersistentVolume

	if err := k8sutil.APIList[*core.PersistentVolumeList](context.Background(), ls.deps.Client.Kubernetes().CoreV1().PersistentVolumes(), meta.ListOptions{}, func(result *core.PersistentVolumeList, err error) error {
		for _, r := range result.Items {
			volumes = append(volumes, r.DeepCopy())
		}

		return nil
	}); err != nil {
		if err != nil {
			return 0, errors.WithStack(err)
		}
	}
	spec := ls.apiObject.Spec
	availableVolumes := 0
	cleanupBeforeTimestamp := time.Now().Add(time.Hour * -24)
	for _, pv := range volumes {
		if pv.Spec.StorageClassName != spec.StorageClass.Name {
			// Not our storage class
			continue
		}

		// We are under deletion
		if pv.DeletionTimestamp != nil {
			// Do not remove object if we are not the owner
			if ls.isOwnerOf(pv) {
				ls.log.Str("name", pv.GetName()).Warn("PV is being deleted")
				if err := ls.inspectPVFinalizer(pv); err != nil {
					ls.log.Str("name", pv.GetName()).Warn("Unable to remove finalizers")
				}
			}
			continue
		}

		switch pv.Status.Phase {
		case core.VolumeAvailable:
			// Is this an old volume?
			if pv.GetObjectMeta().GetCreationTimestamp().Time.Before(cleanupBeforeTimestamp) {
				// Let's clean it up
				if ls.isOwnerOf(pv) {
					// Cleanup this volume
					if features.LocalStorageReclaimPolicyPass().Enabled() {
						ls.removePVObjectWithLog(pv)
					} else {
						ls.log.Str("name", pv.GetName()).Debug("Added PersistentVolume to cleaner")
						ls.pvCleaner.Add(pv)
					}
				} else {
					ls.log.Str("name", pv.GetName()).Debug("PersistentVolume is not owned by us")
					availableVolumes++
				}
			} else {
				availableVolumes++
			}
		case core.VolumeReleased:
			if ls.isOwnerOf(pv) {
				// Cleanup this volume
				if !features.LocalStorageReclaimPolicyPass().Enabled() {
					ls.log.Str("name", pv.GetName()).Debug("Added PersistentVolume to cleaner")
					ls.pvCleaner.Add(pv)
				} else {
					if pv.Spec.PersistentVolumeReclaimPolicy == core.PersistentVolumeReclaimDelete {
						// We have released PV, now delete it
						ls.log.Str("name", pv.GetName()).Info("PV With ReclaimPolicy Delete in state Released found, deleting")
						ls.removePVObjectWithLog(pv)
					}
				}
			} else {
				ls.log.Str("name", pv.GetName()).Debug("PersistentVolume is not owned by us")
			}
		}
	}
	return availableVolumes, nil
}

func (ls *LocalStorage) inspectPVFinalizer(pv *core.PersistentVolume) error {
	currentFinalizers := pv.GetFinalizers()
	if len(currentFinalizers) == 0 {
		// No finalizers, nothing to do
		return nil
	}

	finalizers := make([]string, 0, len(currentFinalizers))

	for _, finalizer := range pv.GetFinalizers() {
		switch finalizer {
		case FinalizerPersistentVolumeCleanup:
			ls.log.Str("name", pv.GetName()).Str("finalizer", FinalizerPersistentVolumeCleanup).Info("Removing finalizer")
			if err := ls.removePVFinalizerPersistentVolumeCleanup(pv); err != nil {
				ls.log.Err(err).Str("name", pv.GetName()).Warn("Unable to remove finalizer")
				finalizers = append(finalizers, finalizer)
			}
		default:
			finalizers = append(finalizers, finalizer)
		}
	}

	// No change in finalizers, all good
	if len(finalizers) == len(currentFinalizers) {
		return nil
	}

	p := patch.NewPatch()
	if len(finalizers) == 0 {
		// Remove them all
		p.Add(patch.ItemRemove(patch.NewPath("metadata", "finalizers")))
	} else {
		p.Add(patch.ItemReplace(patch.NewPath("metadata", "finalizers"), finalizers))
	}

	data, err := p.Marshal()
	if err != nil {
		return err
	}

	if _, err := ls.deps.Client.Kubernetes().CoreV1().PersistentVolumes().Patch(context.Background(), pv.GetName(), types.JSONPatchType, data, meta.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func (ls *LocalStorage) removePVFinalizerPersistentVolumeCleanup(pv *core.PersistentVolume) error {
	if !features.LocalStorageReclaimPolicyPass().Enabled() {
		return nil
	}

	// Find local path
	localSource := pv.Spec.PersistentVolumeSource.Local
	if localSource == nil {
		return errors.WithStack(errors.Newf("PersistentVolume has no local source"))
	}
	localPath := localSource.Path

	// Find client that serves the node
	nodeName := pv.GetAnnotations()[nodeNameAnnotation]
	if nodeName == "" {
		return errors.WithStack(errors.Newf("PersistentVolume has no node-name annotation"))
	}
	client, err := ls.GetClientByNodeName(context.Background(), nodeName)
	if err != nil {
		ls.log.Err(err).Str("node", nodeName).Debug("Failed to get client for node")
		return errors.WithStack(err)
	}

	// Clean volume through client
	ctx := context.Background()
	if err := client.Remove(ctx, localPath); err != nil {
		ls.log.Err(err).
			Str("node", nodeName).
			Str("local-path", localPath).
			Debug("Failed to remove local path")
		return errors.WithStack(err)
	}

	return nil
}

func (ls *LocalStorage) removePVObjectWithLog(pv *core.PersistentVolume) {
	if pv == nil {
		return
	}

	if pv.DeletionTimestamp != nil {
		// Already deleting. nothing to do
		return
	}

	ls.removePVWithLog(pv.GetName(), string(pv.GetUID()))
}

func (ls *LocalStorage) removePVWithLog(name, uid string) {
	if err := ls.removePV(name, uid); err != nil {
		ls.log.Str("name", name).Err(err).Warn("PersistentVolume cannot be removed")
	}
}

func (ls *LocalStorage) removePV(name, uid string) error {
	if err := ls.deps.Client.Kubernetes().CoreV1().PersistentVolumes().Delete(context.Background(), name, meta.DeleteOptions{
		Preconditions: meta.NewUIDPreconditions(uid),
	}); err != nil {
		if apiErrors.IsNotFound(err) {
			// Do not remove if not found
			return nil
		}

		if apiErrors.IsConflict(err) {
			// Do not throw error if uid changed
			return nil
		}

		return err
	}

	return nil
}
