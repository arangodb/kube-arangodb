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
	"fmt"
	"strconv"

	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
)

const (
	annStorageClassIsDefault = "storageclass.kubernetes.io/is-default-class"
)

// StorageClassIsDefault returns true if the given storage class is marked default,
// false otherwise.
func StorageClassIsDefault(sc *v1.StorageClass) bool {
	value, found := sc.GetObjectMeta().GetAnnotations()[annStorageClassIsDefault]
	if !found {
		return false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return boolValue
}

// PatchStorageClassIsDefault changes the default flag of the given storage class.
func PatchStorageClassIsDefault(cli storagev1.StorageV1Interface, name string, isDefault bool) error {
	jsonPatch := fmt.Sprintf(`{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"%v"}}}`, isDefault)
	if _, err := cli.StorageClasses().Patch(name, types.StrategicMergePatchType, []byte(jsonPatch)); err != nil {
		return maskAny(err)
	}
	return nil
}
