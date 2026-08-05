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

package openid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Session_Expires(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var s *Session
		require.True(t, s.Expires().IsZero())
	})

	t.Run("Value", func(t *testing.T) {
		ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		s := &Session{ExpiresAt: meta.NewTime(ts)}
		require.Equal(t, ts, s.Expires())
	})
}

func Test_Session_AsResponse(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var s *Session
		require.Nil(t, s.AsResponse())
	})

	t.Run("Value", func(t *testing.T) {
		s := &Session{Username: "root"}
		r := s.AsResponse()
		require.NotNil(t, r)
		require.Equal(t, "root", r.User)
		require.Nil(t, r.Token)
		require.Empty(t, r.Groups)
	})
}
