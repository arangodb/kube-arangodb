//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

func ParallelProcess[T any](caller func(in T), threads int, in []T) {
	r := ParallelInput(in)

	var wg sync.WaitGroup

	for id := 0; id < threads; id++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for el := range r {
				caller(el)
			}
		}()
	}

	wg.Wait()
}

func ParallelInput[T any](in []T) <-chan T {
	r := make(chan T)

	go func() {
		defer close(r)

		for id := range in {
			r <- in[id]
		}
	}()

	return r
}

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
