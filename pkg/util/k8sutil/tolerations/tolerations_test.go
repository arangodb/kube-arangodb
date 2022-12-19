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

package tolerations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
)

const (
	TolerationValid          = ""
	TolerationUnableToRemove = "Unable to remove toleration"
	TolerationUnableToModify = "Unable to modify toleration"
)

func copyTolerations(in []core.Toleration) []core.Toleration {
	out := make([]core.Toleration, len(in))

	for id := range in {
		in[id].DeepCopyInto(&out[id])

	}

	return out
}

func areTolerationsValid(a, b []core.Toleration) string {
	if len(a) > len(b) {
		return TolerationUnableToRemove
	}

	for id := range a {
		if a[id].Operator != b[id].Operator ||
			a[id].Key != b[id].Key ||
			a[id].Effect != b[id].Effect ||
			a[id].Value != b[id].Value {
			return TolerationUnableToModify
		}
	}

	return TolerationValid
}

func mergeTolerations(t *testing.T, tolerations []core.Toleration, toAdd ...core.Toleration) []core.Toleration {
	return ensureTolerationImmutable(t, tolerations, func(in []core.Toleration) []core.Toleration {
		return MergeTolerationsIfNotFound(tolerations, toAdd)
	}, func(t *testing.T, change string) {
		require.Equal(t, TolerationValid, change)
	})
}

func ensureTolerationImmutable(t *testing.T, tolerations []core.Toleration, mod func(in []core.Toleration) []core.Toleration, check func(t *testing.T, change string)) []core.Toleration {
	param := copyTolerations(tolerations)
	param = mod(param)

	r := areTolerationsValid(tolerations, param)

	check(t, r)
	return param
}

func Test_Tolerations(t *testing.T) {
	var tolerations []core.Toleration

	t.Run("Add initial toleration", func(t *testing.T) {
		tolerations = mergeTolerations(t, tolerations, NewNoExecuteToleration(TolerationKeyNodeNotReady, TolerationDuration{Forever: true}))

		require.Len(t, tolerations, 1)

		require.Nil(t, tolerations[0].TolerationSeconds)
	})

	t.Run("Modify initial toleration", func(t *testing.T) {
		tolerations = mergeTolerations(t, tolerations, NewNoExecuteToleration(TolerationKeyNodeNotReady, TolerationDuration{TimeSpan: 5 * time.Second}))

		require.Len(t, tolerations, 1)

		require.NotNil(t, tolerations[0].TolerationSeconds)
		require.EqualValues(t, 5, *tolerations[0].TolerationSeconds)
	})

	t.Run("Add second toleration", func(t *testing.T) {
		tolerations = mergeTolerations(t, tolerations, NewNoExecuteToleration(TolerationKeyNodeAlphaUnreachable, TolerationDuration{TimeSpan: 5 * time.Second}))

		require.Len(t, tolerations, 2)

		require.NotNil(t, tolerations[0].TolerationSeconds)
		require.EqualValues(t, 5, *tolerations[0].TolerationSeconds)

		require.NotNil(t, tolerations[1].TolerationSeconds)
		require.EqualValues(t, 5, *tolerations[1].TolerationSeconds)
	})

	t.Run("Modify initial toleration again", func(t *testing.T) {
		tolerations = mergeTolerations(t, tolerations, NewNoExecuteToleration(TolerationKeyNodeNotReady, TolerationDuration{TimeSpan: 10 * time.Second}))

		require.Len(t, tolerations, 2)

		require.NotNil(t, tolerations[0].TolerationSeconds)
		require.EqualValues(t, 10, *tolerations[0].TolerationSeconds)

		require.NotNil(t, tolerations[1].TolerationSeconds)
		require.EqualValues(t, 5, *tolerations[1].TolerationSeconds)
	})
}
