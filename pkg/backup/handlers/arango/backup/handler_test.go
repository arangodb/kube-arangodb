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
// Author Adam Janikowski
//

package backup

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
)

func Test_ObjectNotFound(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	i := newItem(operation.Add, "test", "test")

	actions := map[operation.Operation]bool{
		operation.Add:    false,
		operation.Update: false,
		operation.Delete: false,
	}

	// Act
	for operation, shouldFail := range actions {
		t.Run(string(operation), func(t *testing.T) {
			err := handler.Handle(i)

			// Assert
			if shouldFail {
				require.Error(t, err)
				require.True(t, errors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
