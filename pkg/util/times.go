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

package util

import (
	"math"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TimeCompareEqual compares two times, allowing an error of 1s
func TimeCompareEqual(a, b metav1.Time) bool {
	return math.Abs(a.Time.Sub(b.Time).Seconds()) <= 1
}

// TimeCompareEqualPointer compares two times, allowing an error of 1s
func TimeCompareEqualPointer(a, b *metav1.Time) bool {
	if a == nil || b == nil {
		return false
	} else if a == b {
		return true
	}

	return TimeCompareEqual(*a, *b)
}
