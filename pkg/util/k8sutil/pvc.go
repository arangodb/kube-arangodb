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

package k8sutil

import (
	"context"
	"strconv"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

// ExtractStorageResourceRequirement filters resource requirements for Pods.
// Keep reference for backward compatibility
var ExtractStorageResourceRequirement = kresources.ExtractStorageResourceRequirement

// IsPersistentVolumeClaimMarkedForDeletion returns true if the pvc has been marked for deletion.
func IsPersistentVolumeClaimMarkedForDeletion(pvc *core.PersistentVolumeClaim) bool {
	return pvc.DeletionTimestamp != nil
}

// IsPersistentVolumeClaimFileSystemResizePending returns true if the pvc has FileSystemResizePending set to true
func IsPersistentVolumeClaimFileSystemResizePending(pvc *core.PersistentVolumeClaim) bool {
	for _, c := range pvc.Status.Conditions {
		if c.Type == core.PersistentVolumeClaimFileSystemResizePending && c.Status == core.ConditionTrue {
			return true
		}
	}
	return false
}

// IsPersistentVolumeClaimResizing returns true if the pvc has Resizing set to true
func IsPersistentVolumeClaimResizing(pvc *core.PersistentVolumeClaim) bool {
	for _, c := range pvc.Status.Conditions {
		if c.Type == core.PersistentVolumeClaimResizing && c.Status == core.ConditionTrue {
			return true
		}
	}
	return false
}

// CreatePersistentVolumeClaim creates a persistent volume claim with given name and configuration.
// If the pvc already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePersistentVolumeClaim(ctx context.Context, pvcs generic.ModClient[*core.PersistentVolumeClaim], pvcName, deploymentName,
	storageClassName, role string, enforceAntiAffinity bool, resources core.ResourceRequirements,
	vct *core.PersistentVolumeClaim, finalizers []string, owner meta.OwnerReference) error {
	labels := LabelsForDeployment(deploymentName, role)
	volumeMode := core.PersistentVolumeFilesystem
	pvc := &core.PersistentVolumeClaim{
		ObjectMeta: meta.ObjectMeta{
			Name:       pvcName,
			Labels:     labels,
			Finalizers: finalizers,
			Annotations: map[string]string{
				utilConstants.AnnotationEnforceAntiAffinity: strconv.FormatBool(enforceAntiAffinity),
			},
		},
	}
	if vct == nil {
		pvc.Spec = core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			},
			VolumeMode: &volumeMode,
			Resources:  ExtractStorageResourceRequirement(resources),
		}
	} else {
		pvc.Spec = vct.Spec
	}

	if storageClassName != "" {
		pvc.Spec.StorageClassName = &storageClassName
	}
	AddOwnerRefToObject(pvc.GetObjectMeta(), &owner)
	if _, err := pvcs.Create(ctx, pvc, meta.CreateOptions{}); err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.WithStack(err)
	}
	return nil
}
