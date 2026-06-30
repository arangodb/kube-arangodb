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

package policy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_Statement_Validate covers validation of policy statements.
//
// Empty actions/resources lists are intentionally allowed: such a statement
// matches nothing at evaluation time (it grants nothing), so it is accepted rather
// than rejected. Only malformed entries (bad action format, unknown effect) fail.
func Test_Statement_Validate(t *testing.T) {
	t.Run("valid statement", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Actions: Actions{"database:read"}, Resources: Resources{"*"}}
		require.NoError(t, s.Validate())
	})

	t.Run("empty actions list allowed", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Actions: Actions{}, Resources: Resources{"*"}}
		require.NoError(t, s.Validate())
	})

	t.Run("nil actions list allowed", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Resources: Resources{"*"}}
		require.NoError(t, s.Validate())
	})

	t.Run("empty resources list allowed", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Actions: Actions{"database:read"}, Resources: Resources{}}
		require.NoError(t, s.Validate())
	})

	t.Run("fully empty statement allowed", func(t *testing.T) {
		s := Statement{Effect: EffectAllow}
		require.NoError(t, s.Validate())
	})

	t.Run("invalid action format rejected", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Actions: Actions{"database"}, Resources: Resources{"*"}}
		require.Error(t, s.Validate())
	})

	t.Run("invalid effect rejected", func(t *testing.T) {
		s := Statement{Effect: Effect("Maybe"), Actions: Actions{"database:read"}, Resources: Resources{"*"}}
		require.Error(t, s.Validate())
	})

	t.Run("wildcard action and resource accepted", func(t *testing.T) {
		s := Statement{Effect: EffectAllow, Actions: Actions{"*"}, Resources: Resources{"*"}}
		require.NoError(t, s.Validate())
	})

	t.Run("empty statements list allowed", func(t *testing.T) {
		require.NoError(t, Statements{}.Validate())
	})

	t.Run("statements list validates each entry", func(t *testing.T) {
		valid := Statements{
			{Effect: EffectAllow, Actions: Actions{"database:read"}, Resources: Resources{"*"}},
			{Effect: EffectDeny, Actions: Actions{"database:drop"}, Resources: Resources{"production"}},
		}
		require.NoError(t, valid.Validate())

		withInvalid := Statements{
			{Effect: EffectAllow, Actions: Actions{"database:read"}, Resources: Resources{"*"}},
			{Effect: EffectAllow, Actions: Actions{"not-a-valid-action"}, Resources: Resources{"*"}},
		}
		require.Error(t, withInvalid.Validate())
	})
}
