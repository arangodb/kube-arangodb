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

package resources

import (
	"context"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	memberState "github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
)

const (
	// minMemberAge is the minimum duration we expect a member to be created before we remove it because
	// it is not part of a deployment.
	minMemberAge = time.Minute * 10
)

var (
	cleanupRemovedMembersCounters = metrics.MustRegisterCounterVec(metricsComponent, "cleanup_removed_members", "Number of cleanup-removed-members actions", metrics.DeploymentName, metrics.Result)
)

// SyncMembersInCluster sets proper condition for all arangod members that belongs to the deployment.
func (r *Resources) SyncMembersInCluster(ctx context.Context, health memberState.Health) error {
	log := r.log.Str("section", "members")

	if health.Error != nil {
		log.Err(health.Error).Info("Health of the cluster is missing")
		return nil
	}

	// Decide what to do depending on cluster mode
	switch r.context.GetSpec().GetMode() {
	case api.DeploymentModeCluster:
		deploymentName := r.context.GetAPIObject().GetName()
		if err := r.syncMembersInCluster(ctx, health); err != nil {
			cleanupRemovedMembersCounters.WithLabelValues(deploymentName, metrics.Failed).Inc()
			return errors.WithStack(err)
		}
		cleanupRemovedMembersCounters.WithLabelValues(deploymentName, metrics.Success).Inc()
		return nil
	default:
		// Other mode have no concept of cluster in which members can be removed
		return nil
	}
}

// syncMembersInCluster sets proper condition for all arangod members that are part of the cluster.
func (r *Resources) syncMembersInCluster(ctx context.Context, health memberState.Health) error {
	log := r.log.Str("section", "members")
	serverFound := func(id string) bool {
		_, found := health.Members[driver.ServerID(id)]
		return found
	}

	status, lastVersion := r.context.GetStatus()
	updateStatusNeeded := false

	status.Members.ForeachServerInGroups(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			log := log.Str("member", m.ID).Str("role", group.AsRole())
			if serverFound(m.ID) {
				// Member is (still) found, skip it
				if m.Conditions.Update(api.ConditionTypeMemberOfCluster, true, "", "") {
					if err := status.Members.Update(m, group); err != nil {
						log.Err(err).Warn("Failed to update member")
					}
					updateStatusNeeded = true
					log.Debug("Updating MemberOfCluster condition to true")
				}
				continue
			} else if !m.Conditions.IsTrue(api.ConditionTypeMemberOfCluster) {
				if m.Age() < minMemberAge {
					log.Dur("age", m.Age()).Debug("Member is not yet recorded as member of cluster")
					continue
				}
				log.Warn("Member can not be found in cluster")
			} else {
				log.Info("Member is no longer part of the ArangoDB cluster")
			}
		}
		return nil
	}, api.ServerGroupCoordinators, api.ServerGroupDBServers)

	if updateStatusNeeded {
		log.Debug("UpdateStatus needed")

		if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
			log.Err(err).Warn("Failed to update deployment status")
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Resources) EnsureArangoMembers(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	// Create all missing arangomembers
	s, _ := r.context.GetStatus()
	obj := r.context.GetAPIObject()

	if err := s.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, member := range list {
			name := member.ArangoMemberName(r.context.GetAPIObject().GetName(), group)

			c := r.context.WithCurrentArangoMember(name)

			if !c.Exists(ctx) {
				// Create ArangoMember
				obj := &api.ArangoMember{
					ObjectMeta: meta.ObjectMeta{
						Name: name,
						OwnerReferences: []meta.OwnerReference{
							obj.AsOwner(),
						},
					},
					Spec: api.ArangoMemberSpec{
						Group:         group,
						ID:            member.ID,
						DeploymentUID: obj.GetUID(),
					},
				}

				if err := r.context.WithCurrentArangoMember(name).Create(ctx, obj); err != nil {
					return err
				}

				continue
			} else {
				if err := c.Update(ctx, func(m *api.ArangoMember) bool {
					changed := false
					if len(m.OwnerReferences) == 0 {
						m.OwnerReferences = []meta.OwnerReference{
							obj.AsOwner(),
						}
						changed = true
					}

					if m.Spec.DeploymentUID == "" {
						m.Spec.DeploymentUID = obj.GetUID()
						changed = true
					}

					return changed
				}); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := cachedStatus.ArangoMember().V1().Iterate(func(member *api.ArangoMember) error {
		_, g, ok := s.Members.ElementByID(member.Spec.ID)

		if !ok || g != member.Spec.Group {
			// Remove member
			if err := r.context.WithCurrentArangoMember(member.GetName()).Delete(ctx); err != nil {
				return err
			}
		}

		return nil
	}, arangomemberv1.FilterByDeploymentUID(obj.GetUID())); err != nil {
		return err
	}

	return nil
}
