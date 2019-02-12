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
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// prepareAgencyPodTermination checks if the given agency pod is allowed to terminate
// and if so, prepares it for termination.
// It returns nil if the pod is allowed to terminate, an error otherwise.
func (r *Resources) prepareAgencyPodTermination(ctx context.Context, log zerolog.Logger, p *v1.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug().Msg("Pod is already failed, safe to remove agency serving finalizer")
		return nil
	}
	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug().Msg("Entire deployment is being deleted, safe to remove agency serving finalizer")
		return nil
	}

	// Check node the pod is scheduled on
	agentDataWillBeGone := false
	if p.Spec.NodeName != "" {
		node, err := r.context.GetKubeCli().CoreV1().Nodes().Get(p.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get node for member")
			return maskAny(err)
		}
		if node.Spec.Unschedulable {
			agentDataWillBeGone = true
		}
	}

	// Check PVC
	pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(apiObject.GetNamespace())
	pvc, err := pvcs.Get(memberStatus.PersistentVolumeClaimName, metav1.GetOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get PVC for member")
		return maskAny(err)
	}
	if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) {
		agentDataWillBeGone = true
	}

	// Is this a simple pod restart?
	if !agentDataWillBeGone {
		log.Debug().Msg("Pod is just being restarted, safe to terminate agency pod")
		return nil
	}

	// Inspect agency state
	log.Debug().Msg("Agent data will be gone, so we will check agency serving status first")
	ctx = agency.WithAllowNoLeader(ctx)                     // The ID we're checking may be the leader, so ignore situations where all other agents are followers
	ctx, cancel := context.WithTimeout(ctx, time.Second*15) // Force a quick check
	defer cancel()
	agencyConns, err := r.context.GetAgencyClients(ctx, func(id string) bool { return id != memberStatus.ID })
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return maskAny(err)
	}
	if len(agencyConns) == 0 {
		log.Debug().Err(err).Msg("No more remaining agents, we cannot delete this one")
		return maskAny(fmt.Errorf("No more remaining agents"))
	}
	if err := agency.AreAgentsHealthy(ctx, agencyConns); err != nil {
		log.Debug().Err(err).Msg("Remaining agents are not healthy")
		return maskAny(err)
	}

	// Complete agent recovery is needed, since data is already gone or not accessible
	if memberStatus.Conditions.Update(api.ConditionTypeAgentRecoveryNeeded, true, "Data Gone", "") {
		if err := updateMember(memberStatus); err != nil {
			return maskAny(err)
		}
	}
	log.Debug().Msg("Agent is ready to be completely recovered.")

	return nil
}

// prepareDBServerPodTermination checks if the given dbserver pod is allowed to terminate
// and if so, prepares it for termination.
// It returns nil if the pod is allowed to terminate, an error otherwise.
func (r *Resources) prepareDBServerPodTermination(ctx context.Context, log zerolog.Logger, p *v1.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug().Msg("Pod is already failed, safe to remove dbserver pod")
		return nil
	}
	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug().Msg("Entire deployment is being deleted, safe to remove dbserver pod")
		return nil
	}

	// Check node the pod is scheduled on
	dbserverDataWillBeGone := false
	if p.Spec.NodeName != "" {
		node, err := r.context.GetKubeCli().CoreV1().Nodes().Get(p.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get node for member")
			return maskAny(err)
		}
		if node.Spec.Unschedulable {
			dbserverDataWillBeGone = true
		}
	}

	// Check PVC
	pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(apiObject.GetNamespace())
	pvc, err := pvcs.Get(memberStatus.PersistentVolumeClaimName, metav1.GetOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get PVC for member")
		return maskAny(err)
	}
	if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) {
		dbserverDataWillBeGone = true
	}

	// Is this a simple pod restart?
	if !dbserverDataWillBeGone {
		log.Debug().Msg("Pod is just being restarted, safe to remove dbserver pod")
		return nil
	}

	// Inspect cleaned out state
	log.Debug().Msg("DBServer data is being deleted, so we will cleanout the dbserver first")
	c, err := r.context.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return maskAny(err)
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return maskAny(err)
	}
	cleanedOut, err := cluster.IsCleanedOut(ctx, memberStatus.ID)
	if err != nil {
		return maskAny(err)
	}
	if cleanedOut {
		// Cleanout completed
		if memberStatus.Conditions.Update(api.ConditionTypeCleanedOut, true, "CleanedOut", "") {
			if err := updateMember(memberStatus); err != nil {
				return maskAny(err)
			}
		}
		log.Debug().Msg("DBServer is cleaned out.")
		return nil
	}
	// Not cleaned out yet, check member status
	if memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		log.Warn().Msg("Member is already terminated before it could be cleaned out. Not good, but removing dbserver pod because we cannot do anything further")
		// At this point we have to set CleanedOut to true,
		// because we can no longer reason about the state in the agency and
		// bringing back the dbserver again may result in an cleaned out server without us knowing
		memberStatus.Conditions.Update(api.ConditionTypeCleanedOut, true, "Draining server failed", "")
		memberStatus.CleanoutJobID = ""
		if memberStatus.Phase == api.MemberPhaseDrain {
			memberStatus.Phase = api.MemberPhaseCreated
		}
		if err := updateMember(memberStatus); err != nil {
			return maskAny(err)
		}
		return nil
	}
	if memberStatus.Phase == api.MemberPhaseCreated {
		// No cleanout job triggered
		var jobID string
		ctx = driver.WithJobIDResponse(ctx, &jobID)
		// Ensure the cleanout is triggered
		log.Debug().Msg("Server is not yet clean out. Triggering a clean out now")
		if err := cluster.CleanOutServer(ctx, memberStatus.ID); err != nil {
			log.Debug().Err(err).Msg("Failed to clean out server")
			return maskAny(err)
		}
		memberStatus.CleanoutJobID = jobID
		memberStatus.Phase = api.MemberPhaseDrain
		if err := updateMember(memberStatus); err != nil {
			return maskAny(err)
		}
	} else if memberStatus.Phase == api.MemberPhaseDrain {
		// Check the job progress
		agency, err := r.context.GetAgency(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create agency client")
			return maskAny(err)
		}
		jobStatus, err := arangod.CleanoutServerJobStatus(ctx, memberStatus.CleanoutJobID, c, agency)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch cleanout job status")
			return maskAny(err)
		}
		if jobStatus.IsFailed() {
			log.Warn().Str("reason", jobStatus.Reason()).Msg("Cleanout Job failed. Aborting plan")
			// Revert cleanout state
			memberStatus.Phase = api.MemberPhaseCreated
			memberStatus.CleanoutJobID = ""
			return maskAny(fmt.Errorf("Clean out server job failed"))
		}
	}

	return maskAny(fmt.Errorf("Server is not yet cleaned out"))

}
