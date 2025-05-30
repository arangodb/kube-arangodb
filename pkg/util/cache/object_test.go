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

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Object(t *testing.T) {
	var iter int

	obj := NewObject[int](func(ctx context.Context) (int, time.Duration, error) {
		iter++
		return iter, 100 * time.Millisecond, nil
	})

	v, err := obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)

	v, err = obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)

	v, err = obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)

	time.Sleep(200 * time.Millisecond)

	v, err = obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, v)

	time.Sleep(50 * time.Millisecond)

	v, err = obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, v)

	time.Sleep(55 * time.Millisecond)

	v, err = obj.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, v)
}
