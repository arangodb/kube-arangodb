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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_HeadersParse(t *testing.T) {
	t.Run("Simple header", func(t *testing.T) {
		h := ParseHeaders("gzip")

		t.Run("Exists", func(t *testing.T) {
			require.Equal(t, "gzip", h.Accept("gzip"))
		})

		t.Run("Missing", func(t *testing.T) {
			require.Equal(t, "identity", h.Accept("bz"))
		})
	})
	t.Run("Advanced header", func(t *testing.T) {
		h := ParseHeaders("deflate, gzip;q=1.0, *;q=0.5")

		t.Run("Exists", func(t *testing.T) {
			require.Equal(t, "gzip", h.Accept("gzip"))
		})

		t.Run("Not accepted", func(t *testing.T) {
			require.Equal(t, "identity", h.Accept("gz"))
		})

		t.Run("Accepted", func(t *testing.T) {
			require.Equal(t, "br", h.Accept("br"))
		})

		t.Run("MultiAccept - Pick Higher prio", func(t *testing.T) {
			require.Equal(t, "gzip", h.Accept("br", "gzip"))
		})

		t.Run("MultiAccept - Pick Same prio by order", func(t *testing.T) {
			require.Equal(t, "br", h.Accept("br", "compress"))
		})

		t.Run("MultiAccept - Pick Same prio by order", func(t *testing.T) {
			require.Equal(t, "compress", h.Accept("compress", "br"))
		})

		t.Run("MultiAccept - Pick Same prio by order - ignore missing", func(t *testing.T) {
			require.Equal(t, "br", h.Accept("zz", "br"))
		})
	})
}
