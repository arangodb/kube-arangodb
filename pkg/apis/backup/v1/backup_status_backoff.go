//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ArangoBackupStatusBackOff struct {
	Iterations int       `json:"iterations,omitempty"`
	Next       meta.Time `json:"next,omitempty"`
}

func (a *ArangoBackupStatusBackOff) GetIterations() int {
	if a == nil {
		return 0
	}

	if a.Iterations < 0 {
		return 0
	}

	return a.Iterations
}

func (a *ArangoBackupStatusBackOff) GetNext() meta.Time {
	if a == nil {
		return meta.Time{}
	}

	return a.Next
}

func (a *ArangoBackupStatusBackOff) ShouldBackoff(spec *ArangoBackupSpecBackOff) bool {
	return spec == nil || spec.MaxIterations == nil || a.GetIterations() < *spec.MaxIterations
}

func (a *ArangoBackupStatusBackOff) Backoff(spec *ArangoBackupSpecBackOff) *ArangoBackupStatusBackOff {
	if !a.ShouldBackoff(spec) {
		// Do not backoff anymore
		return &ArangoBackupStatusBackOff{
			Iterations: a.GetIterations(),
			Next:       a.GetNext(),
		}
	}

	return &ArangoBackupStatusBackOff{
		Iterations: a.GetIterations() + 1,
		Next:       meta.Time{Time: time.Now().Add(spec.Backoff(a.GetIterations()))},
	}
}
