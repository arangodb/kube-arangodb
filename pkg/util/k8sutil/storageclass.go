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

package k8sutil

import (
	"strconv"
	"time"

	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

var (
	annStorageClassIsDefault = []string{
		// Make sure first entry is the one we'll put in
		"storageclass.kubernetes.io/is-default-class",
		"storageclass.beta.kubernetes.io/is-default-class",
	}
)

// StorageClassIsDefault returns true if the given storage class is marked default,
// false otherwise.
func StorageClassIsDefault(sc *v1.StorageClass) bool {
	for _, key := range annStorageClassIsDefault {
		if value, found := sc.GetObjectMeta().GetAnnotations()[key]; found {
			if boolValue, err := strconv.ParseBool(value); err == nil && boolValue {
				return true
			}
		}
	}
	return false
}

// PatchStorageClassIsDefault changes the default flag of the given storage class.
func PatchStorageClassIsDefault(cli storagev1.StorageV1Interface, name string, isDefault bool) error {
	stcs := cli.StorageClasses()
	op := func() error {
		// Fetch current version of StorageClass
		current, err := stcs.Get(name, metav1.GetOptions{})
		if IsNotFound(err) {
			return retry.Permanent(maskAny(err))
		} else if err != nil {
			return maskAny(err)
		}
		// Tweak annotations
		ann := current.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		for _, key := range annStorageClassIsDefault {
			delete(ann, key)
		}
		ann[annStorageClassIsDefault[0]] = strconv.FormatBool(isDefault)
		current.SetAnnotations(ann)
		// Save StorageClass
		if _, err := stcs.Update(current); IsConflict(err) {
			// StorageClass has been modified since we read it
			return maskAny(err)
		} else if err != nil {
			return retry.Permanent(maskAny(err))
		}
		return nil
	}
	if err := retry.Retry(op, time.Second*15); err != nil {
		return maskAny(err)
	}
	return nil
}
