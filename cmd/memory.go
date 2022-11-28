//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"runtime"
	"syscall"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func monitorMemoryLimit() {
	if memoryLimit.hardLimit == 0 {
		return
	}

	var m runtime.MemStats

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			runtime.ReadMemStats(&m)

			if m.Sys > 1024*1024*memoryLimit.hardLimit {
				if err := syscall.Kill(syscall.Getpid(), syscall.SIGABRT); err != nil {
					panic(err)
				}
			}
		case <-shutdown.Channel():
			return
		}
	}
}
