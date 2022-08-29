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

package conn

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func IsAsyncErrorNotFound(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(asyncErrorNotFound); ok {
		return true
	}

	return IsAsyncErrorNotFound(errors.CauseWithNil(err))
}

func newAsyncErrorNotFound(id string) error {
	return asyncErrorNotFound{
		jobID: id,
	}
}

type asyncErrorNotFound struct {
	jobID string
}

func (a asyncErrorNotFound) Error() string {
	return fmt.Sprintf("Job with ID %s not found", a.jobID)
}

func IsAsyncJobInProgress(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	if v, ok := err.(asyncJobInProgress); ok {
		return v.jobID, true
	}

	return IsAsyncJobInProgress(errors.CauseWithNil(err))
}

func newAsyncJobInProgress(id string) error {
	return asyncJobInProgress{
		jobID: id,
	}
}

type asyncJobInProgress struct {
	jobID string
}

func (a asyncJobInProgress) Error() string {
	return fmt.Sprintf("Job with ID %s in progress", a.jobID)
}
