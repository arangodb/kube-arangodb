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

package storage

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// inspectPVCs queries all PVC's and checks if there is a need to
// build new persistent volumes.
// Returns the PVC's that need a volume.
func (ls *LocalStorage) inspectPVCs() ([]core.PersistentVolumeClaim, error) {
	ns := ls.apiObject.GetNamespace()
	list, err := ls.deps.Client.Kubernetes().CoreV1().PersistentVolumeClaims(ns).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	spec := ls.apiObject.Spec
	var result []core.PersistentVolumeClaim
	for _, pvc := range list.Items {
		if !pvcMatchesStorageClass(pvc, spec.StorageClass.Name, spec.StorageClass.IsDefault) {
			continue
		}
		if !pvcNeedsVolume(pvc) {
			continue
		}
		result = append(result, pvc)
	}
	return result, nil
}

// pvcMatchesStorageClass checks if the given pvc requests a volume
// of the given storage class.
func pvcMatchesStorageClass(pvc core.PersistentVolumeClaim, storageClassName string, isDefault bool) bool {
	scn := pvc.Spec.StorageClassName
	if scn == nil {
		// No storage class specified, default is used
		return isDefault
	}
	return *scn == storageClassName
}

// pvcNeedsVolume checks if the given pvc is in need of a persistent volume.
func pvcNeedsVolume(pvc core.PersistentVolumeClaim) bool {
	return pvc.Status.Phase == core.ClaimPending
}
