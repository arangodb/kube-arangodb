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
	"testing"

	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageClassIsDefault returns true if the given storage class is marked default,
// false otherwise.
func TestStorageClassIsDefault(t *testing.T) {
	tests := []struct {
		StorageClass v1.StorageClass
		IsDefault    bool
	}{
		{v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}, false},
		{v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annStorageClassIsDefault: "false",
				},
			},
		}, false},
		{v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annStorageClassIsDefault: "foo",
				},
			},
		}, false},
		{v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annStorageClassIsDefault: "true",
				},
			},
		}, true},
	}
	for _, test := range tests {
		result := StorageClassIsDefault(&test.StorageClass)
		if result != test.IsDefault {
			t.Errorf("StorageClassIsDefault failed. Expected %v, got %v for %#v", test.IsDefault, result, test.StorageClass)
		}
	}
}
