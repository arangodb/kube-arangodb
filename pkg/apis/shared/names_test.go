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

package shared

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation"
)

func Test_Names(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		require.EqualError(t, ValidateResourceName(""), "Name '' is not a valid resource name")
	})
	t.Run("Pod name is valid", func(t *testing.T) {
		name := CreatePodHostName("the-matrix-db", "arangodb-coordinator", "CRDN-549cznuy")
		require.Empty(t, validation.IsQualifiedName(name))

		name = CreatePodHostName("the-matrix-application-db-instance", "arangodb-coordinator", "CRDN-549cznuy")
		require.Empty(t, validation.IsQualifiedName(name))
	})
}
