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

package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test creation of local storage spec
func TestLocalStorageSpecCreation(t *testing.T) {

	class := StorageClassSpec{"SpecName", true, nil}
	local := LocalStorageSpec{StorageClass: class, LocalPath: []string{""}}
	assert.Error(t, local.Validate())

	class = StorageClassSpec{"spec-name", true, nil}
	local = LocalStorageSpec{StorageClass: class, LocalPath: []string{""}}
	assert.Error(t, local.Validate(), "should fail as the empty sting is not a valid path")

	class = StorageClassSpec{"spec-name", true, nil}
	local = LocalStorageSpec{StorageClass: class, LocalPath: []string{}}
	assert.True(t, IsValidation(local.Validate()))
}

// Test reset of local storage spec
func TestLocalStorageSpecReset(t *testing.T) {
	class := StorageClassSpec{"spec-name", true, nil}
	source := LocalStorageSpec{StorageClass: class, LocalPath: []string{"/a/path", "/another/path"}}
	target := LocalStorageSpec{}
	resetImmutableFieldsResult := source.ResetImmutableFields(&target)
	expected := []string{"storageClass.name", "localPath"}
	assert.Equal(t, expected, resetImmutableFieldsResult)
	assert.Equal(t, source.LocalPath, target.LocalPath)
	assert.Equal(t, source.StorageClass.Name, target.StorageClass.Name)
}
