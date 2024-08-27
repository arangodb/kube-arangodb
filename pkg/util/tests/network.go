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
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ResolveAddress(t *testing.T, addr string) (string, int) {
	ln, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	pr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	addr = pr.IP.String()
	port := pr.Port

	require.NoError(t, ln.Close())
	return addr, port
}

func WaitForAddress(t *testing.T, addr string, port int) {
	tickerT := time.NewTicker(125 * time.Millisecond)
	defer tickerT.Stop()

	timerT := time.NewTimer(5 * time.Second)
	defer timerT.Stop()

	for {
		select {
		case <-timerT.C:
			require.Fail(t, "Timeouted", addr, port)
		case <-tickerT.C:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), 125*time.Millisecond)
			if err != nil {
				continue
			}

			require.NoError(t, conn.Close())

			return
		}
	}
}
