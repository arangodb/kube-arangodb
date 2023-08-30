//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Buffer(t *testing.T) {
	t.Run("Normal cache", func(t *testing.T) {
		up := bytes.NewBuffer(nil)
		b := NewBuffer(4096, up)

		data := make([]byte, 1024)

		c, err := util.WriteAll(b, data)
		require.Len(t, data, c)

		require.NoError(t, err)

		require.Len(t, up.Bytes(), 0)
		require.Len(t, b.Bytes(), 1024)
		require.False(t, b.Truncated())
	})
	t.Run("Full cache", func(t *testing.T) {
		up := bytes.NewBuffer(nil)
		b := NewBuffer(1024, up)

		data := make([]byte, 1024)

		c, err := util.WriteAll(b, data)
		require.Len(t, data, c)

		require.NoError(t, err)

		require.Len(t, up.Bytes(), 0)
		require.Len(t, b.Bytes(), 1024)
		require.False(t, b.Truncated())
	})
	t.Run("Full cache + 1", func(t *testing.T) {
		up := bytes.NewBuffer(nil)
		b := NewBuffer(1024, up)

		data := make([]byte, 1025)

		c, err := util.WriteAll(b, data)
		require.Len(t, data, c)

		require.NoError(t, err)

		require.Len(t, up.Bytes(), 1025)
		require.Len(t, b.Bytes(), 0)
		require.True(t, b.Truncated())
	})
	t.Run("Overflow cache", func(t *testing.T) {
		up := bytes.NewBuffer(nil)
		b := NewBuffer(1024, up)

		data := make([]byte, 2048)

		c, err := util.WriteAll(b, data)
		require.Len(t, data, c)

		require.NoError(t, err)

		require.Len(t, up.Bytes(), 2048)
		require.Len(t, b.Bytes(), 0)
		require.True(t, b.Truncated())
	})
}
