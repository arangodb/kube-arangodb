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
// Author Jan Christoph Uhde
//

package v1alpha

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_LocalStorageSpec_Creation(t *testing.T) {
	var (
		class StorageClassSpec
		local LocalStorageSpec
		err   error
	)

	class = StorageClassSpec{"SpecName", true}
	local = LocalStorageSpec{StorageClass: class, LocalPath: []string{""}}
	err = local.Validate()
	assert.Equal(t, errors.Cause(class.Validate()), errors.Cause(err))

	class = StorageClassSpec{"spec-name", true}
	local = LocalStorageSpec{StorageClass: class, LocalPath: []string{""}} //is this allowed - should the paths be checked?
	err = local.Validate()
	assert.Equal(t, nil, errors.Cause(err))

	class = StorageClassSpec{"spec-name", true}
	local = LocalStorageSpec{StorageClass: class, LocalPath: []string{}}
	err = local.Validate()
	assert.Equal(t, ValidationError, errors.Cause(err)) //path empty
}

func Test_LocalStorageSpec_Reset(t *testing.T) {
	class := StorageClassSpec{"spec-name", true}
	source := LocalStorageSpec{StorageClass: class, LocalPath: []string{"/a/path", "/another/path"}}
	target := LocalStorageSpec{}
	resetImmutableFieldsResult := source.ResetImmutableFields(&target)
	expected := []string{"storageClass.name", "localPath"}
	assert.Equal(t, expected, resetImmutableFieldsResult)
	assert.Equal(t, source.LocalPath, target.LocalPath)
	assert.Equal(t, source.StorageClass.Name, target.StorageClass.Name)
}
