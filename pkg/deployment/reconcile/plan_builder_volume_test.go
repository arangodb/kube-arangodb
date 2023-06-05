//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_MemberConditionTypeMemberVolumeUnschedulableLocalStorageGone(t *testing.T) {
	type testCase struct {
		pv core.PersistentVolumeSpec

		node *core.Node

		result bool
	}

	testCases := map[string]testCase{
		"Non LocalVolume": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					GCEPersistentDisk: &core.GCEPersistentDiskVolumeSource{},
				},
				NodeAffinity: nil,
			},
		},
		"LocalVolume without selectors": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: nil,
			},
		},
		"LocalVolume with partial selectors - NPE#1": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{},
			},
		},
		"LocalVolume with partial selectors - NPE#2": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{},
				},
			},
		},
		"LocalVolume with partial selectors - NPE#3": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{},
					},
				},
			},
		},
		"LocalVolume with partial selectors - NPE#4": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{},
							},
						},
					},
				},
			},
		},
		"LocalVolume with invalid selector key": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{
									{
										Key:      shared.NodeArchAffinityLabel,
										Operator: core.NodeSelectorOpIn,
										Values: []string{
											"node",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"LocalVolume with invalid selector operator": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{
									{
										Key:      shared.TopologyKeyHostname,
										Operator: core.NodeSelectorOpDoesNotExist,
										Values: []string{
											"node",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"LocalVolume with valid selector - existing node": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{
									{
										Key:      shared.TopologyKeyHostname,
										Operator: core.NodeSelectorOpIn,
										Values: []string{
											"node",
										},
									},
								},
							},
						},
					},
				},
			},

			node: &core.Node{
				ObjectMeta: meta.ObjectMeta{
					Name: "node",
				},
			},
		},
		"LocalVolume with valid selector - missing node #1": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{
									{
										Key:      shared.TopologyKeyHostname,
										Operator: core.NodeSelectorOpIn,
										Values: []string{
											"node",
										},
									},
								},
							},
						},
					},
				},
			},

			node: &core.Node{
				ObjectMeta: meta.ObjectMeta{
					Name: "node1",
				},
			},

			result: true,
		},
		"LocalVolume with valid selector - missing node #2": {
			pv: core.PersistentVolumeSpec{
				PersistentVolumeSource: core.PersistentVolumeSource{
					Local: &core.LocalVolumeSource{},
				},
				NodeAffinity: &core.VolumeNodeAffinity{
					Required: &core.NodeSelector{
						NodeSelectorTerms: []core.NodeSelectorTerm{
							{
								MatchExpressions: []core.NodeSelectorRequirement{
									{
										Key:      shared.TopologyKeyHostname,
										Operator: core.NodeSelectorOpIn,
										Values: []string{
											"node",
										},
									},
								},
							},
						},
					},
				},
			},

			result: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			client := kclient.NewFakeClient()

			if tc.node != nil {
				_, err := client.Kubernetes().CoreV1().Nodes().Create(context.Background(), tc.node, meta.CreateOptions{})
				require.NoError(t, err)
			}

			ins := tests.NewInspector(t, client)

			require.Equal(t, tc.result, memberConditionTypeMemberVolumeUnschedulableLocalStorageGone(ins, &core.PersistentVolume{
				Spec: tc.pv,
			}, &core.PersistentVolumeClaim{}))
		})
	}

}
