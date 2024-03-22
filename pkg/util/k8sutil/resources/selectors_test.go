//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Selectors_Labels(t *testing.T) {
	labels := map[string]string{
		"A": "B",
		"C": "D",
		"E": "F",
	}

	t.Run("Match", func(t *testing.T) {
		t.Run("Do not match with nil", func(t *testing.T) {
			require.False(t, SelectLabels(nil, nil))
		})
		t.Run("Match with any", func(t *testing.T) {
			require.True(t, SelectLabels(&meta.LabelSelector{}, nil))
		})
		t.Run("Match with dedicated labels select", func(t *testing.T) {
			require.True(t, SelectLabels(&meta.LabelSelector{
				MatchLabels: map[string]string{
					"A": "B",
				},
			}, labels))
		})
		t.Run("Match with multiple dedicated labels select", func(t *testing.T) {
			require.True(t, SelectLabels(&meta.LabelSelector{
				MatchLabels: map[string]string{
					"A": "B",
					"E": "F",
				},
			}, labels))
		})
		t.Run("Match with mismatch dedicated labels select", func(t *testing.T) {
			require.False(t, SelectLabels(&meta.LabelSelector{
				MatchLabels: map[string]string{
					"A": "B",
					"E": "G",
				},
			}, labels))
		})
	})

	t.Run("Match Expression", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			t.Run("Present", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpExists,
						},
					},
				}, labels))
			})
			t.Run("Missing", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "B",
							Operator: meta.LabelSelectorOpExists,
						},
					},
				}, labels))
			})
		})

		t.Run("Exists", func(t *testing.T) {
			t.Run("Present", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpDoesNotExist,
						},
					},
				}, labels))
			})
			t.Run("Missing", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "B",
							Operator: meta.LabelSelectorOpDoesNotExist,
						},
					},
				}, labels))
			})
		})

		t.Run("In", func(t *testing.T) {
			t.Run("Empty", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpIn,
						},
					},
				}, labels))
			})
			t.Run("Present", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpIn,
							Values: []string{
								"B",
							},
						},
					},
				}, labels))
			})
			t.Run("Present Multiple", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpIn,
							Values: []string{
								"E",
								"Z",
								"B",
							},
						},
					},
				}, labels))
			})
			t.Run("Missing", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "B",
							Operator: meta.LabelSelectorOpIn,
							Values: []string{
								"B",
							},
						},
					},
				}, labels))
			})
		})

		t.Run("NotIn", func(t *testing.T) {
			t.Run("Not Existing", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "Z",
							Operator: meta.LabelSelectorOpNotIn,
						},
					},
				}, labels))
			})
			t.Run("Empty", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpNotIn,
						},
					},
				}, labels))
			})
			t.Run("Present", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpNotIn,
							Values: []string{
								"B",
							},
						},
					},
				}, labels))
			})
			t.Run("Present Multiple", func(t *testing.T) {
				require.False(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpNotIn,
							Values: []string{
								"E",
								"Z",
								"B",
							},
						},
					},
				}, labels))
			})
			t.Run("Missing", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "B",
							Operator: meta.LabelSelectorOpNotIn,
							Values: []string{
								"B",
							},
						},
					},
				}, labels))
			})
			t.Run("Missing Value", func(t *testing.T) {
				require.True(t, SelectLabels(&meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "A",
							Operator: meta.LabelSelectorOpNotIn,
							Values: []string{
								"R",
								"Z",
								"D",
							},
						},
					},
				}, labels))
			})
		})
	})
}
