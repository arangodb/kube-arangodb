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

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func ParallelProcessErr[T any](caller func(in T) error, threads int, in []T) error {
	errs := ParallelProcessOutput[T, error](caller, threads, in)

	return errors.Errors(errs...)
}

func ParallelProcessOutput[T, O any](caller func(in T) O, threads int, in []T) []O {
	r := ParallelInput(IntInput(len(in)))
	ret := make([]O, len(in))
	var wg sync.WaitGroup

	for id := 0; id < threads; id++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for id := range r {
				ret[id] = caller(in[id])
			}
		}()
	}

	wg.Wait()

	return ret
}

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

func IntInput(count int) []int {
	var r = make([]int, count)
	for i := 0; i < count; i++ {
		r[i] = i
	}
	return r
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
