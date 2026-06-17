//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package topology

import (
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type Mapping map[int]api.List

func GetTopologyMapping(status api.DeploymentStatus) (Mapping, error) {
	m := Mapping{}

	for _, member := range status.Members.AsList() {
		if !status.Topology.IsTopologyOwned(member.Member.Topology) {
			continue
		}

		if member.Member.Topology.Label == "" {
			continue
		}

		for k, v := range m {
			if k == member.Member.Topology.Zone {
				continue
			}

			if v.Contains(member.Member.Topology.Label) {
				return nil, errors.Errorf("Multi assignment")
			}
		}

		m[member.Member.Topology.Zone] = m[member.Member.Topology.Zone].Add(member.Member.Topology.Label).Unique().Sort()
	}

	return m, nil
}

func GetTopologyAffinityRules(name string, status api.DeploymentStatus, group api.ServerGroup, member api.MemberStatus) core.Affinity {
	var a = core.Affinity{
		NodeAffinity: &core.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{},
		},
		PodAffinity:     &core.PodAffinity{},
		PodAntiAffinity: &core.PodAntiAffinity{},
	}

	if !status.Topology.Enabled() || !status.Topology.IsTopologyOwned(member.Topology) || member.Topology.ID != status.Topology.ID {
		return a
	}

	coreLabels := []meta.LabelSelectorRequirement{
		{
			Key:      k8sutil.LabelKeyApp,
			Operator: meta.LabelSelectorOpIn,
			Values:   []string{k8sutil.AppName},
		},
		{
			Key:      k8sutil.LabelKeyArangoDeployment,
			Operator: meta.LabelSelectorOpIn,
			Values:   []string{name},
		},
	}

	topologyOriented := mergeTopologyLabels(coreLabels, meta.LabelSelectorRequirement{
		Key:      k8sutil.LabelKeyArangoTopology,
		Operator: meta.LabelSelectorOpIn,
		Values:   []string{string(status.Topology.ID)},
	})

	var nodeSelectorRequirements []core.NodeSelectorRequirement

	members := status.Members.AsList().Filter(func(a api.DeploymentStatusMemberElement) bool {
		return status.Topology.IsTopologyOwned(a.Member.Topology) && a.Member.Topology.Zone == member.Topology.Zone
	}).Sort(func(a, b api.DeploymentStatusMemberElement) bool {
		return a.Member.CreatedAt.Before(&b.Member.CreatedAt)
	})

	if member.IsInitialized && member.Topology.Label != "" {
		nodeSelectorRequirements = append(nodeSelectorRequirements, core.NodeSelectorRequirement{
			Key:      status.Topology.Label,
			Operator: core.NodeSelectorOpIn,
			Values: []string{
				member.Topology.Label,
			},
		})
	} else {
		if len(members) == 0 || members[0].Member.ID == member.ID {
			// Select zone
			a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				core.WeightedPodAffinityTerm{
					Weight: 100,
					PodAffinityTerm: core.PodAffinityTerm{
						LabelSelector: &meta.LabelSelector{
							MatchExpressions: mergeTopologyLabels(topologyOriented, meta.LabelSelectorRequirement{
								Key:      k8sutil.LabelKeyArangoZone,
								Operator: meta.LabelSelectorOpIn,
								Values:   []string{fmt.Sprintf("%d", member.Topology.Zone)},
							}),
						},
						TopologyKey: status.Topology.Label,
					},
				},
				core.WeightedPodAffinityTerm{
					Weight: 50,
					PodAffinityTerm: core.PodAffinityTerm{
						LabelSelector: &meta.LabelSelector{
							MatchExpressions: mergeTopologyLabels(coreLabels, meta.LabelSelectorRequirement{
								Key:      k8sutil.LabelKeyArangoZone,
								Operator: meta.LabelSelectorOpExists,
							}),
						},
						TopologyKey: status.Topology.Label,
					},
				},
			)
		} else {
			// Wait for first member
			a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
				core.PodAffinityTerm{
					LabelSelector: &meta.LabelSelector{
						MatchExpressions: mergeTopologyLabels(topologyOriented, meta.LabelSelectorRequirement{
							Key:      k8sutil.LabelKeyArangoZone,
							Operator: meta.LabelSelectorOpIn,
							Values:   []string{fmt.Sprintf("%d", member.Topology.Zone)},
						}, meta.LabelSelectorRequirement{
							Key:      k8sutil.LabelKeyArangoScheduled,
							Operator: meta.LabelSelectorOpExists,
						}),
					},
					TopologyKey: status.Topology.Label,
				},
			)
		}
	}

	// Get all zones
	var otherZones []string
	var usedLabels []string

	for i := 0; i < status.Topology.Size; i++ {
		if member.Topology.Zone == i {
			continue
		}

		otherZones = append(otherZones, fmt.Sprintf("%d", i))
		usedLabels = append(usedLabels, status.Topology.Zones[i].Labels...)
	}

	// Do not schedule on zones which are already in use
	if len(otherZones) > 0 {
		a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			core.PodAffinityTerm{
				LabelSelector: &meta.LabelSelector{
					MatchExpressions: mergeTopologyLabels(topologyOriented, meta.LabelSelectorRequirement{
						Key:      k8sutil.LabelKeyArangoZone,
						Operator: meta.LabelSelectorOpIn,
						Values:   otherZones,
					}),
				},
				TopologyKey: status.Topology.Label,
			},
		)
	}

	// Schedule only on zones allocated for this member
	if len(usedLabels) > 0 {
		nodeSelectorRequirements = append(nodeSelectorRequirements, core.NodeSelectorRequirement{
			Key:      status.Topology.Label,
			Operator: core.NodeSelectorOpNotIn,
			Values:   usedLabels,
		})
	}

	if len(nodeSelectorRequirements) > 0 {
		a.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []core.NodeSelectorTerm{
			{
				MatchExpressions: nodeSelectorRequirements,
			},
		}
	}

	return a
}

func mergeTopologyLabels(c []meta.LabelSelectorRequirement, extension ...meta.LabelSelectorRequirement) []meta.LabelSelectorRequirement {
	ret := make([]meta.LabelSelectorRequirement, len(c)+len(extension))

	for id := range c {
		c[id].DeepCopyInto(&ret[id])
	}
	for id := range extension {
		extension[id].DeepCopyInto(&ret[id+len(c)])
	}

	return ret
}
