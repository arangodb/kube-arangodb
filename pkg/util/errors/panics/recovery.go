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

package panics

func recoverPanic(skipFrames int, in func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newPanicError(r, GetStack(skipFrames))
		}
	}()

	return in()
}

func recoverPanicO1[O1 any](skipFrames int, in func() (O1, error)) (o1 O1, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newPanicError(r, GetStack(skipFrames))

			logger.Err(err).Error("Panic: %+v", err)
		}
	}()

	return in()
}

func Recover(in func() error) (err error) {
	return recoverPanic(4, in)
}

func RecoverO1[O1 any](in func() (O1, error)) (O1, error) {
	return recoverPanicO1(4, in)
}
