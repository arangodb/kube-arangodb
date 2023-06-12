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

package nctx

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func counterRequestReadMock(t *testing.T, ctx context.Context, data io.Reader) int {
	data = WithRequestReadBytes(ctx, data)
	dz, err := io.ReadAll(data)
	require.NoError(t, err)
	return len(dz)
}

func Test_Counter_RequestRead(t *testing.T) {
	data := make([]byte, 64)

	var c Counter

	t.Run("Read without wrapper", func(t *testing.T) {
		require.EqualValues(t, 64, counterRequestReadMock(t, context.Background(), bytes.NewReader(data)))
		require.EqualValues(t, 0, c.Get())
	})

	t.Run("Read with wrapper", func(t *testing.T) {
		require.EqualValues(t, 64, counterRequestReadMock(t, c.WithRequestReadBytes(context.Background()), bytes.NewReader(data)))
		require.EqualValues(t, 64, c.Get())
	})
}
