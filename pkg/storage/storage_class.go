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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	storageClassProvisioner = api.SchemeGroupVersion.Group + "/localstorage"
)

// ensureStorageClass creates a storage class for the given local storage.
// If such a class already exists, the create is ignored.
func (l *LocalStorage) ensureStorageClass(apiObject *api.ArangoLocalStorage) error {
	log := l.deps.Log
	spec := apiObject.Spec.StorageClass
	bindingMode := v1.VolumeBindingWaitForFirstConsumer
	reclaimPolicy := corev1.PersistentVolumeReclaimRetain
	sc := &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		ReclaimPolicy:     &reclaimPolicy,
		VolumeBindingMode: &bindingMode,
		Provisioner:       storageClassProvisioner,
	}
	// Note: We do not attach the StorageClass to the apiObject (OwnerRef) because many
	// ArangoLocalStorage resource may use the same StorageClass.
	cli := l.deps.Client.Kubernetes().StorageV1()
	if _, err := cli.StorageClasses().Create(context.Background(), sc, metav1.CreateOptions{}); k8sutil.IsAlreadyExists(err) {
		log.Debug().
			Str("storageclass", sc.GetName()).
			Msg("StorageClass already exists")
	} else if err != nil {
		log.Debug().Err(err).
			Str("storageclass", sc.GetName()).
			Msg("Failed to create StorageClass")
		return errors.WithStack(err)
	} else {
		log.Debug().
			Str("storageclass", sc.GetName()).
			Msg("StorageClass created")
	}

	if apiObject.Spec.StorageClass.IsDefault {
		// UnMark current default (if any)
		list, err := cli.StorageClasses().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Debug().Err(err).Msg("Listing StorageClasses failed")
			return errors.WithStack(err)
		}
		for _, scX := range list.Items {
			if !k8sutil.StorageClassIsDefault(&scX) || scX.GetName() == sc.GetName() {
				continue
			}
			// Mark storage class as non-default
			if err := k8sutil.PatchStorageClassIsDefault(cli, scX.GetName(), false); err != nil {
				log.Debug().
					Err(err).
					Str("storageclass", scX.GetName()).
					Msg("Failed to mark StorageClass as not-default")
				return errors.WithStack(err)
			}
			log.Debug().
				Str("storageclass", scX.GetName()).
				Msg("Marked StorageClass as not-default")
		}

		// Mark StorageClass default
		if err := k8sutil.PatchStorageClassIsDefault(cli, sc.GetName(), true); err != nil {
			log.Debug().
				Err(err).
				Str("storageclass", sc.GetName()).
				Msg("Failed to mark StorageClass as default")
			return errors.WithStack(err)
		}
		log.Debug().
			Str("storageclass", sc.GetName()).
			Msg("Marked StorageClass as default")
	}

	return nil
}
