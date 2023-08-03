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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Action_Marshal(t *testing.T) {
	var a Action

	data, err := json.Marshal(a)
	require.NoError(t, err)
	require.Equal(t, `{"id":"","type":"","creationTime":null}`, string(data))
}

func Test_Action_Equal(t *testing.T) {
	a := Action{
		ID:       "9ktKQsBAqe5Ra8ZJ",
		SetID:    "",
		Type:     ActionTypeAddMember,
		MemberID: "",
		Group:    2,
		CreationTime: meta.Time{
			Time: time.Date(2023, time.August, 1, 20, 6, 53, 0, time.UTC),
		},
		StartTime:    nil,
		Reason:       "",
		Image:        "",
		Params:       nil,
		Locals:       nil,
		TaskID:       "",
		Architecture: "",
		Progress:     "",
	}
	b := Action{
		ID:       "9ktKQsBAqe5Ra8ZJ",
		SetID:    "",
		Type:     ActionTypeAddMember,
		MemberID: "",
		Group:    2,
		CreationTime: meta.Time{
			Time: time.Date(2023, time.August, 1, 20, 6, 53, 0, time.UTC),
		},
		StartTime:    nil,
		Reason:       "",
		Image:        "",
		Params:       nil,
		Locals:       nil,
		TaskID:       "",
		Architecture: "",
		Progress:     "",
	}

	require.True(t, a.Equal(a))
	require.True(t, a.Equal(b))
	require.True(t, b.Equal(a))
	require.True(t, b.Equal(b))

	now := time.Now()
	a.StartTime = &meta.Time{Time: now}
	require.True(t, a.Equal(a))
	require.False(t, a.Equal(b))
	require.False(t, b.Equal(a))
	require.True(t, b.Equal(b))
}
