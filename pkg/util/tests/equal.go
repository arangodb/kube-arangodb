//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

//go:build testing

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func EqualPointers[A any](t *testing.T, a, b *A) {
	if a == nil {
		require.Nil(t, b, "Both objects expected to be nil")
	} else {
		require.NotNil(t, b, "Both objects expected not to be nil")
		require.EqualValues(t, *a, *b)
	}
}
