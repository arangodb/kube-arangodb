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

package k8sutil

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

// CreateAffinity creates pod anti-affinity for the given role.
// role contains the name of the role to configure any-affinity with.
// affinityWithRole contains the role to configure affinity with.
func CreateAffinity(deploymentName, role string, required bool, affinityWithRole string) *core.Affinity {
	a := &core.Affinity{
		NodeAffinity: &core.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
				NodeSelectorTerms: []core.NodeSelectorTerm{
					{
						MatchExpressions: []core.NodeSelectorRequirement{
							{
								Key:      shared.NodeArchAffinityLabel,
								Operator: "In",
								Values:   []string{"amd64"},
							},
						},
					},
				},
			},
		},
		PodAntiAffinity: &core.PodAntiAffinity{},
	}
	labels := LabelsForDeployment(deploymentName, role)
	labelSelector := &meta.LabelSelector{
		MatchLabels: labels,
	}
	if required {
		a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, core.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   shared.TopologyKeyHostname,
		})
	} else {
		a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
			Weight: 1,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   shared.TopologyKeyHostname,
			},
		})
	}
	if affinityWithRole != "" {
		a.PodAffinity = &core.PodAffinity{}
		labelSelector := &meta.LabelSelector{
			MatchLabels: LabelsForDeployment(deploymentName, affinityWithRole),
		}
		if required {
			a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   shared.TopologyKeyHostname,
			})
		} else {
			a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
				Weight: 1,
				PodAffinityTerm: core.PodAffinityTerm{
					LabelSelector: labelSelector,
					TopologyKey:   shared.TopologyKeyHostname,
				},
			})
		}
	}

	return a
}
