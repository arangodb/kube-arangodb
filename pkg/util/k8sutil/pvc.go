//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package k8sutil

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

// PersistentVolumeClaimInterface has methods to work with PersistentVolumeClaim resources.
type PersistentVolumeClaimInterface interface {
	Create(*v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error)
	Get(name string, options metav1.GetOptions) (*v1.PersistentVolumeClaim, error)
}

// IsPersistentVolumeClaimMarkedForDeletion returns true if the pvc has been marked for deletion.
func IsPersistentVolumeClaimMarkedForDeletion(pvc *v1.PersistentVolumeClaim) bool {
	return pvc.DeletionTimestamp != nil
}

// IsPersistentVolumeClaimFileSystemResizePending returns true if the pvc has FileSystemResizePending set to true
func IsPersistentVolumeClaimFileSystemResizePending(pvc *v1.PersistentVolumeClaim) bool {
	for _, c := range pvc.Status.Conditions {
		if c.Type == v1.PersistentVolumeClaimFileSystemResizePending && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// CreatePersistentVolumeClaimName returns the name of the persistent volume claim for a member with
// a given id in a deployment with a given name.
func CreatePersistentVolumeClaimName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + stripArangodPrefix(id)
}

// ExtractStorageResourceRequirement filters resource requirements for Pods.
func ExtractStorageResourceRequirement(resources v1.ResourceRequirements) v1.ResourceRequirements {

	filterStorage := func(list v1.ResourceList) v1.ResourceList {
		newlist := make(v1.ResourceList)
		for k, v := range list {
			if k != v1.ResourceStorage && k != "iops" {
				continue
			}
			newlist[k] = v
		}
		return newlist
	}

	return v1.ResourceRequirements{
		Limits:   filterStorage(resources.Limits),
		Requests: filterStorage(resources.Requests),
	}
}

// CreatePersistentVolumeClaim creates a persistent volume claim with given name and configuration.
// If the pvc already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePersistentVolumeClaim(pvcs PersistentVolumeClaimInterface, pvcName, deploymentName, ns, storageClassName, role string, enforceAntiAffinity bool, resources v1.ResourceRequirements, vct *v1.PersistentVolumeClaim, finalizers []string, owner metav1.OwnerReference) error {
	labels := LabelsForDeployment(deploymentName, role)
	volumeMode := v1.PersistentVolumeFilesystem
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:       pvcName,
			Labels:     labels,
			Finalizers: finalizers,
			Annotations: map[string]string{
				constants.AnnotationEnforceAntiAffinity: strconv.FormatBool(enforceAntiAffinity),
			},
		},
	}
	if vct == nil {
		pvc.Spec = v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
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
	if _, err := pvcs.Create(pvc); err != nil && !IsAlreadyExists(err) {
		return maskAny(err)
	}
	return nil
}
