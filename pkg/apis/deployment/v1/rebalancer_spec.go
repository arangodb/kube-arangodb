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

package v1

type ArangoDeploymentRebalancerSpec struct {
	Enabled *bool `json:"enabled"`

	ParallelMoves *int `json:"parallelMoves,omitempty"`

	Readers *ArangoDeploymentRebalancerReadersSpec `json:"readers,omitempty"`

	Optimizers *ArangoDeploymentRebalancerOptimizersSpec `json:"optimizers,omitempty"`
}

func (a *ArangoDeploymentRebalancerSpec) IsEnabled() bool {
	if a == nil {
		return false
	}

	if a.Enabled == nil {
		return true
	}

	return *a.Enabled
}

func (a *ArangoDeploymentRebalancerSpec) GetParallelMoves(d int) int {
	if !a.IsEnabled() {
		return d
	}

	if a == nil || a.ParallelMoves == nil {
		return d
	}

	return *a.ParallelMoves
}

type ArangoDeploymentRebalancerReadersSpec struct {
	Count *bool `json:"count,omitempty"`
}

func (a *ArangoDeploymentRebalancerReadersSpec) IsCountEnabled() bool {
	if a == nil || a.Count == nil {
		return false
	}

	return *a.Count
}

type ArangoDeploymentRebalancerOptimizersSpec struct {
	Leader *bool `json:"leader,omitempty"`
}

func (a *ArangoDeploymentRebalancerOptimizersSpec) IsLeaderEnabled() bool {
	if a == nil || a.Leader == nil {
		return true
	}

	return *a.Leader
}
