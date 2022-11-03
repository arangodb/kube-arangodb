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

package k8sutil

import (
	"reflect"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

type Informer interface {
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}

func WaitForInformers(stop <-chan struct{}, timeout time.Duration, informers ...Informer) {
	done := make(chan struct{})

	go func() {
		defer close(done)

		started := make(chan struct{})

		go func() {
			defer close(started)

			var wg sync.WaitGroup

			for id := range informers {
				wg.Add(1)

				go func(id int) {
					defer wg.Done()
					informers[id].WaitForCacheSync(stop)
				}(id)
			}

			wg.Wait()
		}()

		select {
		case <-started:
		case <-timer.After(timeout):
		}
	}()

	<-done
}
