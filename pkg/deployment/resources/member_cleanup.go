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

package resources

import (
	"context"
	"time"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	// minMemberAge is the minimum duration we expect a member to be created before we remove it because
	// it is not part of a deployment.
	minMemberAge = time.Minute * 10
)

// CleanupRemovedMembers removes all arangod members that are no longer part of ArangoDB deployment.
func (r *Resources) CleanupRemovedMembers() error {
	// Decide what to do depending on cluster mode
	switch r.context.GetSpec().GetMode() {
	case api.DeploymentModeCluster:
		if err := r.cleanupRemovedClusterMembers(); err != nil {
			return maskAny(err)
		}
		return nil
	default:
		// Other mode have no concept of cluster in which members can be removed
		return nil
	}
}

// cleanupRemovedClusterMembers removes all arangod members that are no longer part of the cluster.
func (r *Resources) cleanupRemovedClusterMembers() error {
	log := r.log
	ctx := context.Background()

	// Ask cluster for its health
	client, err := r.context.GetDatabaseClient(ctx)
	if err != nil {
		return maskAny(err)
	}
	c, err := client.Cluster(ctx)
	if err != nil {
		return maskAny(err)
	}
	h, err := c.Health(ctx)
	if err != nil {
		return maskAny(err)
	}

	serverFound := func(id string) bool {
		_, found := h.Health[driver.ServerID(id)]
		return found
	}

	// For over all members that can be removed
	status, lastVersion := r.context.GetStatus()
	updateStatusNeeded := false
	var podNamesToRemove, pvcNamesToRemove []string
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		if group != api.ServerGroupCoordinators && group != api.ServerGroupDBServers {
			// We're not interested in these other groups
			return nil
		}
		for _, m := range list {
			if serverFound(m.ID) {
				// Member is (still) found, skip it
				if m.Conditions.Update(api.ConditionTypeMemberOfCluster, true, "", "") {
					status.Members.Update(m, group)
					updateStatusNeeded = true
				}
				continue
			} else if !m.Conditions.IsTrue(api.ConditionTypeMemberOfCluster) {
				// Member is not yet recorded as member of cluster
				if m.Age() < minMemberAge {
					continue
				}
				log.Info().Str("member", m.ID).Str("role", group.AsRole()).Msg("Member has never been part of the cluster for a long time. Removing it.")
			} else {
				// Member no longer part of cluster, remove it
				log.Info().Str("member", m.ID).Str("role", group.AsRole()).Msg("Member is no longer part of the ArangoDB cluster. Removing it.")
			}
			status.Members.RemoveByID(m.ID, group)
			updateStatusNeeded = true
			// Remove Pod & PVC (if any)
			if m.PodName != "" {
				podNamesToRemove = append(podNamesToRemove, m.PodName)
			}
			if m.PersistentVolumeClaimName != "" {
				pvcNamesToRemove = append(pvcNamesToRemove, m.PersistentVolumeClaimName)
			}
		}
		return nil
	})

	if updateStatusNeeded {
		if err := r.context.UpdateStatus(status, lastVersion); err != nil {
			return maskAny(err)
		}
	}

	for _, podName := range podNamesToRemove {
		if err := r.context.DeletePod(podName); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Str("pod", podName).Msg("Failed to remove obsolete pod")
		}
	}
	for _, pvcName := range pvcNamesToRemove {
		if err := r.context.DeletePvc(pvcName); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Str("pvc", pvcName).Msg("Failed to remove obsolete PVC")
		}
	}

	return nil
}
