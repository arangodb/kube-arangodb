//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createAffinity creates pod anti-affinity for the given role.
// role contains the name of the role to configure any-affinity with.
// affinityWithRole contains the role to configure affinity with.
func createAffinity(deploymentName, role string, required bool, affinityWithRole string) *v1.Affinity {
	a := &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{},
	}
	labels := LabelsForDeployment(deploymentName, role)
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	if required {
		a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, v1.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   TopologyKeyHostname,
		})
	} else {
		a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, v1.WeightedPodAffinityTerm{
			Weight: 1,
			PodAffinityTerm: v1.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   TopologyKeyHostname,
			},
		})
	}
	if affinityWithRole != "" {
		a.PodAffinity = &v1.PodAffinity{}
		labelSelector := &metav1.LabelSelector{
			MatchLabels: LabelsForDeployment(deploymentName, affinityWithRole),
		}
		if required {
			a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, v1.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   TopologyKeyHostname,
			})
		} else {
			a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, v1.WeightedPodAffinityTerm{
				Weight: 1,
				PodAffinityTerm: v1.PodAffinityTerm{
					LabelSelector: labelSelector,
					TopologyKey:   TopologyKeyHostname,
				},
			})
		}
	}
	return a
}
