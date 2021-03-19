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
	"context"
	"strconv"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
)

var (
	annStorageClassIsDefault = "storageclass.kubernetes.io/is-default-class"
)

// StorageClassIsDefault returns true if the given storage class is marked default,
// false otherwise.
func StorageClassIsDefault(sc *v1.StorageClass) bool {
	if value, found := sc.GetObjectMeta().GetAnnotations()[annStorageClassIsDefault]; found {
		if boolValue, err := strconv.ParseBool(value); err == nil && boolValue {
			return true
		}
	}
	return false
}

// PatchStorageClassIsDefault changes the default flag of the given storage class.
func PatchStorageClassIsDefault(cli storagev1.StorageV1Interface, name string, isDefault bool) error {
	stcs := cli.StorageClasses()
	op := func() error {
		// Fetch current version of StorageClass
		current, err := stcs.Get(context.Background(), name, metav1.GetOptions{})
		if IsNotFound(err) {
			return retry.Permanent(errors.WithStack(err))
		} else if err != nil {
			return errors.WithStack(err)
		}
		// Tweak annotations
		ann := current.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		ann[annStorageClassIsDefault] = strconv.FormatBool(isDefault)
		current.SetAnnotations(ann)

		// Save StorageClass
		if _, err := stcs.Update(context.Background(), current, metav1.UpdateOptions{}); IsConflict(err) {
			// StorageClass has been modified since we read it
			return errors.WithStack(err)
		} else if err != nil {
			return retry.Permanent(errors.WithStack(err))
		}
		return nil
	}
	if err := retry.Retry(op, time.Second*15); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
