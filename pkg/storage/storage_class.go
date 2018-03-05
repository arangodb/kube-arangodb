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
// Author Ewout Prangsma
//

package storage

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/storage/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// ensureStorageClass creates a storage class for the given local storage.
// If such a class already exists, the create is ignored.
func (l *LocalStorage) ensureStorageClass(apiObject *api.ArangoLocalStorage) error {
	spec := apiObject.Spec.StorageClass
	bindingMode := v1.VolumeBindingWaitForFirstConsumer
	reclaimPolicy := corev1.PersistentVolumeReclaimRetain
	sc := &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		ReclaimPolicy:     &reclaimPolicy,
		VolumeBindingMode: &bindingMode,
	}
	if _, err := l.deps.KubeCli.StorageV1().StorageClasses().Create(sc); !k8sutil.IsAlreadyExists(err) && err != nil {
		return maskAny(err)
	}
	// TODO make default (if needed)

	return nil
}
