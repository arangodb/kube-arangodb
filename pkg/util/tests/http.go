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

package tests

import (
	"context"
	"fmt"
	"net"
	goHttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
)

func NewHTTPServer(ctx context.Context, t *testing.T, mods ...util.ModEP1[goHttp.Server, context.Context]) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	pr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)

	addr, port := pr.IP.String(), pr.Port

	server, err := http.NewServer(ctx, mods...)
	require.NoError(t, err)

	closer := server.Async(ctx, ln)

	WaitForTCPPort(addr, port).WithContextTimeoutT(t, ctx, 10*time.Second, 125*time.Millisecond)

	go func() {
		require.NoError(t, closer())
	}()

	return fmt.Sprintf("%s:%d", addr, port)
}
