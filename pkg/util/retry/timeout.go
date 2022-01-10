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

package retry

import "time"

func Interrput() error {
	return interrupt{}
}

type interrupt struct {
}

func (i interrupt) Error() string {
	return "interrupt"
}

func IsInterrupt(err error) bool {
	_, ok := err.(interrupt)
	return ok
}

type TimeoutError struct {
}

func (i TimeoutError) Error() string {
	return "timeout"
}

func NewTimeout(t Timeout) Timeout {
	return t
}

type Timeout func() error

func (t Timeout) Timeout(interval, timeout time.Duration) error {
	timeoutI := time.NewTimer(timeout)
	defer timeoutI.Stop()

	intervalI := time.NewTicker(interval)
	defer intervalI.Stop()

	for {
		err := t()

		if err != nil {
			if IsInterrupt(err) {
				return nil
			}

			return err
		}

		select {
		case <-timeoutI.C:
			return TimeoutError{}
		case <-intervalI.C:
			continue
		}
	}
}
