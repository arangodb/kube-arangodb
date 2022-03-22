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

package util

import "sync"

// RunParallel runs actions parallelly throttling them to the given maximum number.
func RunParallel(max int, actions ...func() error) error {
	c, close := ParallelThread(max)
	defer close()

	errors := make([]error, len(actions))

	var wg sync.WaitGroup

	wg.Add(len(actions))
	for id, i := range actions {
		go func(id int, action func() error) {
			defer func() {
				c <- struct{}{}
				wg.Done()
			}()
			<-c

			errors[id] = action()
		}(id, i)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func ParallelThread(max int) (chan struct{}, func()) {
	c := make(chan struct{}, max)

	for i := 0; i < max; i++ {
		c <- struct{}{}
	}

	return c, func() {
		close(c)
		for range c {
		}
	}
}
