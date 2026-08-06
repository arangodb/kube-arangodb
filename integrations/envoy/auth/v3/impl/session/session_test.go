//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_session_Key(t *testing.T) {
	t.Run("Nil returns empty", func(t *testing.T) {
		var s *session
		require.Equal(t, "", s.GetKey())
	})

	t.Run("Get returns what Set stored", func(t *testing.T) {
		s := &session{}
		s.SetKey("my-key")
		require.Equal(t, "my-key", s.GetKey())
	})
}

func Test_session_Expires(t *testing.T) {
	t.Run("Nil returns zero", func(t *testing.T) {
		var s *session
		require.True(t, s.Expires().IsZero())
	})

	t.Run("Value", func(t *testing.T) {
		ts := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
		s := &session{ExpiresAt: meta.NewTime(ts)}
		require.Equal(t, ts, s.Expires())
	})
}
