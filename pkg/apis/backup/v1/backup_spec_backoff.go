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

import "time"

type ArangoBackupSpecBackOff struct {
	// MinDelay defines minimum delay in seconds. Default to 30
	MinDelay *int `json:"min_delay,omitempty"`
	// MaxDelay defines maximum delay in seconds. Default to 600
	MaxDelay *int `json:"max_delay,omitempty"`
	// Iterations defines number of iterations before reaching MaxDelay. Default to 5
	Iterations *int `json:"iterations,omitempty"`
	// MaxIterations defines maximum number of iterations after backoff will be disabled. Default to nil (no limit)
	MaxIterations *int `json:"max_iterations,omitempty"`
}

func (a *ArangoBackupSpecBackOff) GetMaxDelay() int {
	if a == nil || a.MaxDelay == nil {
		return 600
	}

	v := *a.MaxDelay

	if v < 0 {
		return 0
	}

	return v
}

func (a *ArangoBackupSpecBackOff) GetMinDelay() int {
	if a == nil || a.MinDelay == nil {
		return 30
	}

	v := *a.MinDelay

	if v < 0 {
		return 0
	}

	if m := a.GetMaxDelay(); m < v {
		return m
	}

	return v
}

func (a *ArangoBackupSpecBackOff) GetIterations() int {
	if a == nil || a.Iterations == nil {
		return 5
	}

	v := *a.Iterations

	if v < 1 {
		return 1
	}

	return v
}

func (a *ArangoBackupSpecBackOff) Backoff(iteration int) time.Duration {
	if maxIterations := a.GetIterations(); maxIterations <= iteration {
		return time.Duration(a.GetMaxDelay()) * time.Second
	} else {
		min, max := a.GetMinDelay(), a.GetMaxDelay()

		if min == max {
			return time.Duration(min) * time.Second
		}

		return time.Duration(min+int(float64(iteration)/float64(maxIterations)*float64(max-min))) * time.Second
	}
}
