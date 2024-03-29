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

package backup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ObjectNotFound(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	i := tests.NewItem(t, operation.Add, tests.NewMetaObject[*backupApi.ArangoBackup](t, "none", "none"))

	actions := map[operation.Operation]bool{
		operation.Add:    false,
		operation.Update: false,
		operation.Delete: false,
	}

	// Act
	for operation, shouldFail := range actions {
		t.Run(string(operation), func(t *testing.T) {
			err := handler.Handle(context.Background(), i)

			// Assert
			if shouldFail {
				require.Error(t, err)
				require.True(t, apiErrors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
