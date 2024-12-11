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
	"time"
)

func WaitForTCPPort(addr string, port int) Timeout {
	return func() error {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), time.Second)
		if err != nil {
			return nil
		}

		if err := conn.Close(); err != nil {
			return nil
		}

		return Interrupt()
	}
}
