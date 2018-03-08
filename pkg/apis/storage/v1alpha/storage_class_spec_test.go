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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StorageClassSpec_Creation(t *testing.T) {
	storageClassSpec := StorageClassSpec{}
	assert.True(t, nil != storageClassSpec.Validate())

	storageClassSpec = StorageClassSpec{Name: "TheSpecName", IsDefault: true} // no upper-case allowed
	assert.True(t, nil != storageClassSpec.Validate())

	storageClassSpec = StorageClassSpec{"the-spec-name", true}
	assert.Equal(t, nil, storageClassSpec.Validate())

	storageClassSpec = StorageClassSpec{} // this is invalid because it was not created with a proper name
	storageClassSpec.SetDefaults("foo")   // here the Name is fixed
	assert.Equal(t, nil, storageClassSpec.Validate())
}

func Test_StorageClassSpec_ResetImmutableFileds(t *testing.T) {
	specSource := StorageClassSpec{"source", true}
	specTarget := StorageClassSpec{"target", true}

	assert.Equal(t, "target", specTarget.Name)
	rv := specSource.ResetImmutableFields("fieldPrefix-", &specTarget)
	assert.Equal(t, "fieldPrefix-name", strings.Join(rv[:], ", "))
	assert.Equal(t, "source", specTarget.Name)
}
