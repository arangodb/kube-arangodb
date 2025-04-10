//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

package metrics

import (
	"fmt"
	"net"
	goHttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func StartHTTP(t *testing.T, mux *goHttp.ServeMux) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port

	server := &goHttp.Server{}
	server.Handler = mux

	closed := make(chan struct{})
	go func() {
		defer close(closed)

		if err := server.Serve(listener); err != nil {
			if !errors.Is(err, goHttp.ErrServerClosed) {
				require.NoError(t, err)
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)

	return fmt.Sprintf("http://127.0.0.1:%d", port), func() {
		require.NoError(t, server.Close())
		<-closed
	}
}
