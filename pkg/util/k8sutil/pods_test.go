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
// Author Ewout Prangsma
//

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

// TestIsPodReady tests IsPodReady.
func TestIsPodReady(t *testing.T) {
	assert.False(t, IsPodReady(&v1.Pod{}))
	assert.False(t, IsPodReady(&v1.Pod{
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	}))
	assert.True(t, IsPodReady(&v1.Pod{
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}))
}

// TestIsPodFailed tests IsPodFailed.
func TestIsPodFailed(t *testing.T) {
	assert.False(t, IsPodFailed(&v1.Pod{}))
	assert.False(t, IsPodFailed(&v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}))
	assert.True(t, IsPodFailed(&v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}))
}

// TestIsPodSucceeded tests IsPodSucceeded.
func TestIsPodSucceeded(t *testing.T) {
	assert.False(t, IsPodSucceeded(&v1.Pod{}))
	assert.False(t, IsPodSucceeded(&v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}))
	assert.True(t, IsPodSucceeded(&v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}))
}
