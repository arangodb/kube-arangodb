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

package k8sutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	conflictError = apierrors.NewConflict(schema.GroupResource{"groupName", "resourceName"}, "something", os.ErrInvalid)
	existsError   = apierrors.NewAlreadyExists(schema.GroupResource{"groupName", "resourceName"}, "something")
	invalidError  = apierrors.NewInvalid(schema.GroupKind{"groupName", "kindName"}, "something", field.ErrorList{})
	notFoundError = apierrors.NewNotFound(schema.GroupResource{"groupName", "resourceName"}, "something")
)

func TestIsAlreadyExists(t *testing.T) {
	assert.False(t, IsAlreadyExists(conflictError))
	assert.False(t, IsAlreadyExists(maskAny(invalidError)))
	assert.True(t, IsAlreadyExists(existsError))
	assert.True(t, IsAlreadyExists(maskAny(existsError)))
}

func TestIsConflict(t *testing.T) {
	assert.False(t, IsConflict(existsError))
	assert.False(t, IsConflict(maskAny(invalidError)))
	assert.True(t, IsConflict(conflictError))
	assert.True(t, IsConflict(maskAny(conflictError)))
}

func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(invalidError))
	assert.False(t, IsNotFound(maskAny(invalidError)))
	assert.True(t, IsNotFound(notFoundError))
	assert.True(t, IsNotFound(maskAny(notFoundError)))
}
