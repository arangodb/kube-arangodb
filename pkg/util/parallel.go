//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

func RunParallel(max int, actions ...func() error) error {
	c := make(chan int, max)
	errors := make([]error, len(actions))
	defer func() {
		close(c)
		for range c {
		}
	}()

	for i := 0; i < max; i++ {
		c <- 0
	}

	var wg sync.WaitGroup

	for id, i := range actions {
		wg.Add(1)

		go func(id int, action func() error) {
			defer func() {
				c <- 0
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
