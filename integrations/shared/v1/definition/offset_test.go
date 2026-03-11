//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package definition

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Offset(t *testing.T) {
	items := make([]int, 1024)
	for i := range items {
		items[i] = i
	}

	t.Run("Default", func(t *testing.T) {
		res, nitems := Paginate(nil, items)
		require.Len(t, nitems, DefaultPageSize)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 1, res.GetPage())
		require.EqualValues(t, items[0:DefaultPageSize], nitems)
	})

	t.Run("Empty", func(t *testing.T) {
		res, nitems := Paginate(&OffsetRequest{}, items)
		require.Len(t, nitems, DefaultPageSize)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 1, res.GetPage())
		require.EqualValues(t, items[0:DefaultPageSize], nitems)
	})

	t.Run("Full", func(t *testing.T) {
		res, nitems := Paginate(&OffsetRequest{ItemsPerPage: util.NewType[int32](102400)}, items)
		require.Len(t, nitems, 1024)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 1, res.GetPage())
		require.EqualValues(t, items, nitems)
	})

	t.Run("Page 2", func(t *testing.T) {
		res, nitems := Paginate(&OffsetRequest{Page: util.NewType[int32](2)}, items)
		require.Len(t, nitems, DefaultPageSize)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 2, res.GetPage())
		require.EqualValues(t, items[DefaultPageSize:DefaultPageSize*2], nitems)
	})

	t.Run("Page -2", func(t *testing.T) {
		res, nitems := Paginate(&OffsetRequest{Page: util.NewType[int32](-2)}, items)
		require.Len(t, nitems, 0)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 0, res.GetPage())
	})

	t.Run("Page 555", func(t *testing.T) {
		res, nitems := Paginate(&OffsetRequest{Page: util.NewType[int32](555)}, items)
		require.Len(t, nitems, 0)
		require.EqualValues(t, len(items), res.GetItems())
		require.EqualValues(t, 555, res.GetPage())
	})
}
