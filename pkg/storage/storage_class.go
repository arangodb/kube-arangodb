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

	storage "k8s.io/api/storage/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

var (
	storageClassProvisioner = api.SchemeGroupVersion.Group + "/localstorage"
)

// ensureStorageClass creates a storage class for the given local storage.
// If such a class already exists, the create is ignored.
func (l *LocalStorage) ensureStorageClass(apiObject *api.ArangoLocalStorage) error {
	spec := apiObject.Spec.StorageClass
	bindingMode := storage.VolumeBindingWaitForFirstConsumer
	sc := &storage.StorageClass{
		ObjectMeta: meta.ObjectMeta{
			Name: spec.Name,
		},
		ReclaimPolicy:     util.NewType(apiObject.Spec.StorageClass.GetReclaimPolicy()),
		VolumeBindingMode: &bindingMode,
		Provisioner:       storageClassProvisioner,
	}
	// Note: We do not attach the StorageClass to the apiObject (OwnerRef) because many
	// ArangoLocalStorage resource may use the same StorageClass.
	cli := l.deps.Client.Kubernetes().StorageV1()
	if _, err := cli.StorageClasses().Create(context.Background(), sc, meta.CreateOptions{}); kerrors.IsAlreadyExists(err) {
		l.log.
			Str("storageclass", sc.GetName()).
			Debug("StorageClass already exists")
	} else if err != nil {
		l.log.Err(err).
			Str("storageclass", sc.GetName()).
			Debug("Failed to create StorageClass")
		return errors.WithStack(err)
	} else {
		l.log.
			Str("storageclass", sc.GetName()).
			Debug("StorageClass created")
	}

	if apiObject.Spec.StorageClass.IsDefault {
		// UnMark current default (if any)
		list, err := cli.StorageClasses().List(context.Background(), meta.ListOptions{})
		if err != nil {
			l.log.Err(err).Debug("Listing StorageClasses failed")
			return errors.WithStack(err)
		}
		for _, scX := range list.Items {
			if !k8sutil.StorageClassIsDefault(&scX) || scX.GetName() == sc.GetName() {
				continue
			}
			// Mark storage class as non-default
			if err := k8sutil.PatchStorageClassIsDefault(cli, scX.GetName(), false); err != nil {
				l.log.
					Err(err).
					Str("storageclass", scX.GetName()).
					Debug("Failed to mark StorageClass as not-default")
				return errors.WithStack(err)
			}
			l.log.
				Str("storageclass", scX.GetName()).
				Debug("Marked StorageClass as not-default")
		}

		// Mark StorageClass default
		if err := k8sutil.PatchStorageClassIsDefault(cli, sc.GetName(), true); err != nil {
			l.log.
				Err(err).
				Str("storageclass", sc.GetName()).
				Debug("Failed to mark StorageClass as default")
			return errors.WithStack(err)
		}
		l.log.
			Str("storageclass", sc.GetName()).
			Debug("Marked StorageClass as default")
	}

	return nil
}
