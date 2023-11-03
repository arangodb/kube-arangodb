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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
)

func planBuilderScaleDownFilter(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatus, error) {
	return NewScaleFilter(context, status, group, in).
		Filter(planBuilderScaleDownSelectMarkedToRemove).
		Filter(planBuilderScaleDownSelectCleanedOutCondition).
		Filter(planBuilderScaleDownCleanedServers).
		Filter(planBuilderScaleDownToBeCleanedServers).
		Select(planBuilderScaleDownTopologyMissing).
		Select(planBuilderScaleDownTopologyAwarenessMember).
		Filter(planBuilderScaleDownFilterByPriority).
		Select(planBuilderScaleDownLowestShards).
		Get()
}

func planBuilderScaleDownSelectMarkedToRemove(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error) {
	r := make(api.MemberStatusList, 0, len(in))

	for _, el := range in {
		if el.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
			r = append(r, el)
		}
	}

	return r, len(r) > 0, nil
}

func planBuilderScaleDownFilterByPriority(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error) {
	items := map[string]int{}

	for _, el := range in {
		cache, ok := context.ACS().ClusterCache(el.ClusterID)
		if !ok {
			items[el.ID] = 0
			continue
		}
		am, ok := cache.ArangoMember().V1().GetSimple(el.ArangoMemberName(context.GetName(), group))
		if !ok {
			items[el.ID] = 0
			continue
		}

		items[el.ID] = am.Spec.GetDeletionPriority()
	}

	max := 0

	for _, priority := range items {
		if priority > max {
			max = priority
		}
	}

	if max == 0 {
		return nil, false, nil
	}

	r := make(api.MemberStatusList, 0, len(in))

	for _, el := range in {
		priority, ok := items[el.ID]
		if ok {
			if priority == max {
				r = append(r, el)
			}
		}
	}

	return r, len(r) > 0, nil
}

func planBuilderScaleDownSelectCleanedOutCondition(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error) {
	r := make(api.MemberStatusList, 0, len(in))

	for _, el := range in {
		if el.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			r = append(r, el)
		}
	}

	return r, len(r) > 0, nil
}

func planBuilderScaleDownCleanedServers(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error) {
	if group != api.ServerGroupDBServers {
		return nil, false, nil
	}

	if agency, ok := context.GetAgencyCache(); ok {
		r := make(api.MemberStatusList, 0, len(in))

		for _, el := range in {
			if agency.Target.CleanedServers.Contains(state.Server(el.ID)) {
				r = append(r, el)
			}
		}

		return r, len(r) > 0, nil
	}

	return nil, false, nil
}

func planBuilderScaleDownToBeCleanedServers(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error) {
	if group != api.ServerGroupDBServers {
		return nil, false, nil
	}

	if agency, ok := context.GetAgencyCache(); ok {
		r := make(api.MemberStatusList, 0, len(in))

		for _, el := range in {
			if agency.Target.ToBeCleanedServers.Contains(state.Server(el.ID)) {
				r = append(r, el)
			}
		}

		return r, len(r) > 0, nil
	}

	return nil, false, nil
}

func planBuilderScaleDownLowestShards(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatus, bool, error) {
	if group != api.ServerGroupDBServers {
		return api.MemberStatus{}, false, nil
	}

	if agency, ok := context.GetAgencyCache(); ok {
		dbServersShards := agency.ShardsByDBServers()

		for _, member := range in {
			if _, ok := dbServersShards[state.Server(member.ID)]; !ok {
				// member is not in agency cache, so it has no shards
				return member, true, nil
			}
		}

		var resultServer state.Server = ""
		var resultShards int

		for server, shards := range dbServersShards {
			// init first server as result
			if resultServer == "" {
				resultServer = server
				resultShards = shards
			} else if shards < resultShards {
				resultServer = server
				resultShards = shards
			}
		}

		for _, member := range in {
			if member.ID == string(resultServer) {
				return member, true, nil
			}
		}
	}

	return api.MemberStatus{}, false, nil
}
