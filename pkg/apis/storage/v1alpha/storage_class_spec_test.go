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
// Author Jan Christoph Uhde
//

package v1alpha

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test creation of storage class spec
func TestStorageClassSpecCreation(t *testing.T) {
	storageClassSpec := StorageClassSpec{}
	assert.Error(t, storageClassSpec.Validate(), "empty name name is not allowed")

	storageClassSpec = StorageClassSpec{Name: "TheSpecName", IsDefault: true}
	assert.Error(t, storageClassSpec.Validate(), "upper case letters are not allowed in resources")

	storageClassSpec = StorageClassSpec{"the-spec-name", true}
	assert.NoError(t, storageClassSpec.Validate())

	storageClassSpec = StorageClassSpec{} // no proper name -> invalid
	storageClassSpec.SetDefaults("foo")   // name is fixed -> vaild
	assert.NoError(t, storageClassSpec.Validate())
}

// test reset of storage class spec
func TestStorageClassSpecResetImmutableFileds(t *testing.T) {
	specSource := StorageClassSpec{"source", true}
	specTarget := StorageClassSpec{"target", true}

	assert.Equal(t, "target", specTarget.Name)
	rv := specSource.ResetImmutableFields("fieldPrefix-", &specTarget)
	assert.Equal(t, "fieldPrefix-name", strings.Join(rv, ", "))
	assert.Equal(t, "source", specTarget.Name)
}
