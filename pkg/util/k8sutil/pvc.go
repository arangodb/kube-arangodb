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

package k8sutil

import (
	"context"
	"strconv"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

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

// ExtractStorageResourceRequirement filters resource requirements for Pods.
func ExtractStorageResourceRequirement(resources core.ResourceRequirements) core.ResourceRequirements {

	filterStorage := func(list core.ResourceList) core.ResourceList {
		newlist := make(core.ResourceList)
		for k, v := range list {
			if k != core.ResourceStorage && k != "iops" {
				continue
			}
			newlist[k] = v
		}
		return newlist
	}

	return core.ResourceRequirements{
		Limits:   filterStorage(resources.Limits),
		Requests: filterStorage(resources.Requests),
	}
}

// CreatePersistentVolumeClaim creates a persistent volume claim with given name and configuration.
// If the pvc already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePersistentVolumeClaim(ctx context.Context, pvcs persistentvolumeclaimv1.ModInterface, pvcName, deploymentName,
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
				constants.AnnotationEnforceAntiAffinity: strconv.FormatBool(enforceAntiAffinity),
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
