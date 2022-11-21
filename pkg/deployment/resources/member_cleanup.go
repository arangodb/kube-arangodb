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
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
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

	status := r.context.GetStatus()
	updateStatusNeeded := false

	for _, e := range status.Members.AsListInGroups(api.ServerGroupCoordinators, api.ServerGroupDBServers) {
		m := e.Member
		group := e.Group
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

	if updateStatusNeeded {
		log.Debug("UpdateStatus needed")

		if err := r.context.UpdateStatus(ctx, status); err != nil {
			log.Err(err).Warn("Failed to update deployment status")
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Resources) EnsureArangoMembers(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	// Create all missing arangomembers
	s := r.context.GetStatus()
	obj := r.context.GetAPIObject()

	for _, e := range s.Members.AsList() {
		name := e.Member.ArangoMemberName(r.context.GetAPIObject().GetName(), e.Group)

		_, ok := cachedStatus.ArangoMember().V1().GetSimple(name)

		if !ok {
			// Create ArangoMember
			obj := &api.ArangoMember{
				ObjectMeta: meta.ObjectMeta{
					Name: name,
					OwnerReferences: []meta.OwnerReference{
						obj.AsOwner(),
					},
				},
				Spec: api.ArangoMemberSpec{
					Group:         e.Group,
					ID:            e.Member.ID,
					DeploymentUID: obj.GetUID(),
				},
			}

			nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
			defer c()

			if _, err := cachedStatus.ArangoMemberModInterface().V1().Create(nctx, obj, meta.CreateOptions{}); err != nil {
				return err
			}

			continue
		} else {
			if err := inspectorInterface.WithArangoMemberUpdate(ctx, cachedStatus, name, func(in *api.ArangoMember) (bool, error) {
				changed := false
				if len(in.OwnerReferences) == 0 {
					in.OwnerReferences = []meta.OwnerReference{
						obj.AsOwner(),
					}
					changed = true
				}

				if in.Spec.DeploymentUID == "" {
					in.Spec.DeploymentUID = obj.GetUID()
					changed = true
				}

				return changed, nil
			}); err != nil {
				return err
			}
		}
	}

	if err := cachedStatus.ArangoMember().V1().Iterate(func(member *api.ArangoMember) error {
		_, g, ok := s.Members.ElementByID(member.Spec.ID)

		if !ok || g != member.Spec.Group {
			// Remove member
			nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
			defer c()

			if err := cachedStatus.ArangoMemberModInterface().V1().Delete(nctx, member.Name, meta.DeleteOptions{}); err != nil {
				return err
			}
		}

		return nil
	}, arangomemberv1.FilterByDeploymentUID(obj.GetUID())); err != nil {
		return err
	}

	return nil
}
