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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageInfoList(t *testing.T) {
	var list ImageInfoList

	_, found := list.GetByImage("notfound")
	assert.False(t, found)
	_, found = list.GetByImageID("id-notfound")
	assert.False(t, found)

	list.AddOrUpdate(ImageInfo{
		Image:           "foo",
		ImageID:         "foo-ID",
		ArangoDBVersion: "1.3.4",
	})
	assert.Len(t, list, 1)

	_, found = list.GetByImage("foo")
	assert.True(t, found)
	_, found = list.GetByImageID("foo-ID")
	assert.True(t, found)
}
