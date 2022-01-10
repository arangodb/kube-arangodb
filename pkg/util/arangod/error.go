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

package arangod

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

var (
	KeyNotFoundError = errors.New("Key not found")
)

// IsKeyNotFound returns true if the given error is (or is caused by) a KeyNotFoundError.
func IsKeyNotFound(err error) bool {
	return errors.Cause(err) == KeyNotFoundError
}

// NotLeaderError indicates the response of an agent when it is
// not the leader of the agency.
type NotLeaderError struct {
	Leader string // Endpoint of the current leader
}

// Error implements error.
func (e NotLeaderError) Error() string {
	return "not the leader"
}

// IsNotLeader returns true if the given error is (or is caused by) a NotLeaderError.
func IsNotLeader(err error) (string, bool) {
	nlErr, ok := err.(NotLeaderError)
	if ok {
		return nlErr.Leader, true
	}
	return "", false
}
