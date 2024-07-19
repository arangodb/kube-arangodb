//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func resetConfig() {
	configuration = newConfiguration()
}

func execCommand(t *testing.T, args ...string) configurationObject {
	config := newConfiguration()

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	require.NoError(t, config.Init(cmd))

	cmd.SetArgs(args)

	require.NoError(t, cmd.Execute())

	return config
}

func Test_ClientSettings(t *testing.T) {
	t.Run("Ensure nil function is handled", func(t *testing.T) {
		defer resetConfig()

		require.NotNil(t, RoundTripper())
	})

	t.Run("Ensure default settings", func(t *testing.T) {
		defer resetConfig()

		def := newConfiguration()

		evaluated := execCommand(t)

		require.Equal(t, def, evaluated)
	})

	t.Run("Ensure default settings overriden", func(t *testing.T) {
		defer resetConfig()

		def := newConfiguration()

		evaluated := execCommand(t, "--http1.keep-alive=false")

		require.NotEqual(t, def, evaluated)

		evaluated.TransportKeepAlive = true
		require.Equal(t, def, evaluated)
	})

	t.Run("Ensure normal client", func(t *testing.T) {
		evaluated := execCommand(t)

		transport := Transport(evaluated.DefaultTransport)

		c, ok := transport.(*http.Transport)
		require.True(t, ok)

		require.Equal(t, c.DisableKeepAlives, !evaluated.TransportKeepAlive)
		require.Equal(t, c.IdleConnTimeout, evaluated.TransportIdleConnTimeout)
	})

	t.Run("Ensure short client", func(t *testing.T) {
		evaluated := execCommand(t)

		transport := Transport(evaluated.ShortTransport)

		c, ok := transport.(*http.Transport)
		require.True(t, ok)

		require.Equal(t, c.DisableKeepAlives, !evaluated.TransportKeepAlive)
		require.Equal(t, c.IdleConnTimeout, evaluated.TransportIdleConnTimeoutShort)
	})

	t.Run("Ensure normal client with mods", func(t *testing.T) {
		evaluated := execCommand(t, "--http1.transport.idle-conn-timeout=1h")

		transport := Transport(evaluated.DefaultTransport)

		c, ok := transport.(*http.Transport)
		require.True(t, ok)

		require.Equal(t, c.DisableKeepAlives, !evaluated.TransportKeepAlive)
		require.Equal(t, c.IdleConnTimeout, evaluated.TransportIdleConnTimeout)
		require.Equal(t, c.IdleConnTimeout, time.Hour)
	})

	t.Run("Ensure short client with mods", func(t *testing.T) {
		evaluated := execCommand(t, "--http1.transport.idle-conn-timeout-short=6h")

		transport := Transport(evaluated.ShortTransport)

		c, ok := transport.(*http.Transport)
		require.True(t, ok)

		require.Equal(t, c.DisableKeepAlives, !evaluated.TransportKeepAlive)
		require.Equal(t, c.IdleConnTimeout, evaluated.TransportIdleConnTimeoutShort)
		require.Equal(t, c.IdleConnTimeout, 6*time.Hour)
	})
}
