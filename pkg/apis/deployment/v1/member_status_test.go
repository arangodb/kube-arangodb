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

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestMemberStatusRecentTerminations tests the functions related to MemberStatus.RecentTerminations.
func TestMemberStatusRecentTerminations(t *testing.T) {
	relTime := func(delta time.Duration) metav1.Time {
		return metav1.Time{Time: time.Now().Add(delta)}
	}

	s := MemberStatus{}
	assert.Equal(t, 0, s.RecentTerminationsSince(time.Now().Add(-time.Hour)))
	assert.Equal(t, 0, s.RemoveTerminationsBefore(time.Now()))

	s.RecentTerminations = []metav1.Time{metav1.Now()}
	assert.Equal(t, 1, s.RecentTerminationsSince(time.Now().Add(-time.Minute)))
	assert.Equal(t, 0, s.RecentTerminationsSince(time.Now().Add(time.Minute)))
	assert.Equal(t, 0, s.RemoveTerminationsBefore(time.Now().Add(-time.Hour)))

	s.RecentTerminations = []metav1.Time{relTime(-time.Hour), relTime(-time.Minute), relTime(time.Minute)}
	assert.Equal(t, 3, s.RecentTerminationsSince(time.Now().Add(-time.Hour*2)))
	assert.Equal(t, 2, s.RecentTerminationsSince(time.Now().Add(-time.Minute*2)))
	assert.Equal(t, 2, s.RemoveTerminationsBefore(time.Now()))
	assert.Len(t, s.RecentTerminations, 1)
}

// TestMemberStatusIsNotReadySince tests the functions related to MemberStatus.IsNotReadySince.
func TestMemberStatusIsNotReadySince(t *testing.T) {
	s := MemberStatus{
		CreatedAt: metav1.Now(),
	}
	assert.False(t, s.IsNotReadySince(time.Now().Add(-time.Hour)))

	s.CreatedAt.Time = time.Now().Add(-time.Hour)
	assert.False(t, s.IsNotReadySince(time.Now().Add(-2*time.Hour)))
	assert.True(t, s.IsNotReadySince(time.Now().Add(-(time.Hour - time.Minute))))

	s.CreatedAt = metav1.Now()
	s.Conditions.Update(ConditionTypeReady, true, "", "")
	assert.False(t, s.IsNotReadySince(time.Now().Add(-time.Minute)))
	assert.False(t, s.IsNotReadySince(time.Now().Add(time.Minute)))

	s.Conditions.Update(ConditionTypeReady, false, "", "")
	assert.False(t, s.IsNotReadySince(time.Now().Add(-time.Minute)))
	assert.True(t, s.IsNotReadySince(time.Now().Add(time.Minute)))
}
