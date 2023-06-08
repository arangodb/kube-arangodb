//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// prepareAgencyPodTermination checks if the given agency pod is allowed to terminate
// and if so, prepares it for termination.
// It returns nil if the pod is allowed to terminate, an error otherwise.
func (r *Resources) prepareAgencyPodTermination(p *core.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	log := r.log.Str("section", "pod")

	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug("Pod is already failed, safe to remove agency serving finalizer")
		return nil
	}
	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug("Entire deployment is being deleted, safe to remove agency serving finalizer")
		return nil
	}

	// Check node the pod is scheduled on. Only if not in namespaced scope
	agentDataWillBeGone := false
	if nodes, err := r.context.ACS().CurrentClusterCache().Node().V1(); err == nil {
		if !r.context.GetScope().IsNamespaced() && p.Spec.NodeName != "" {
			node, ok := nodes.GetSimple(p.Spec.NodeName)
			if !ok {
				log.Warn("Node not found")
			} else if node.Spec.Unschedulable {
				agentDataWillBeGone = true
			}
		}
	}

	// Check PVC
	pvc, ok := r.context.ACS().CurrentClusterCache().PersistentVolumeClaim().V1().GetSimple(memberStatus.PersistentVolumeClaim.GetName())
	if !ok {
		log.Warn("Failed to get PVC for member")
		return errors.Newf("Failed to get PVC for member")
	}
	if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) {
		agentDataWillBeGone = true
	}

	// Is this a simple pod restart?
	if !agentDataWillBeGone {
		log.Debug("Pod is just being restarted, safe to terminate agency pod")
		return nil
	}

	// Inspect agency state
	log.Debug("Agent data will be gone, so we will check agency serving status first")

	agencyHealth, ok := r.context.GetAgencyHealth()
	if !ok {
		log.Debug("Agency health fetch failed")
		return errors.Newf("Agency health fetch failed")
	}
	if err := agencyHealth.Healthy(); err != nil {
		log.Err(err).Debug("Agency is not healthy. Cannot delete this one")
		return errors.WithStack(errors.Newf("Agency is not healthy"))
	}
	// Complete agent recovery is needed, since data is already gone or not accessible
	if memberStatus.Conditions.Update(api.ConditionTypeAgentRecoveryNeeded, true, "Data Gone", "") {
		if err := updateMember(memberStatus); err != nil {
			return errors.WithStack(err)
		}
	}
	log.Debug("Agent is ready to be completely recovered.")

	return nil
}

// prepareDBServerPodTermination checks if the given dbserver pod is allowed to terminate
// and if so, prepares it for termination.
// It returns nil if the pod is allowed to terminate, an error otherwise.
func (r *Resources) prepareDBServerPodTermination(ctx context.Context, p *core.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	log := r.log.Str("section", "pod")

	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug("Pod is already failed, safe to remove dbserver pod")
		return nil
	}

	// If pod is not member of cluster, do nothing
	if !memberStatus.Conditions.IsTrue(api.ConditionTypeMemberOfCluster) {
		log.Debug("Pod is not member of cluster")
		return nil
	}

	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug("Entire deployment is being deleted, safe to remove dbserver pod")
		return nil
	}

	// Check node the pod is scheduled on
	dbserverDataWillBeGone := false
	if nodes, err := r.context.ACS().CurrentClusterCache().Node().V1(); err == nil {
		node, ok := nodes.GetSimple(p.Spec.NodeName)
		if !ok {
			log.Warn("Node not found")
		} else if node.Spec.Unschedulable {
			if !r.context.GetSpec().IsNetworkAttachedVolumes() {
				dbserverDataWillBeGone = true
			}
		}
	}

	// Check PVC
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	pvc, err := r.context.ACS().CurrentClusterCache().PersistentVolumeClaim().V1().Read().Get(ctxChild, memberStatus.PersistentVolumeClaim.GetName(), meta.GetOptions{})
	if err != nil {
		log.Err(err).Warn("Failed to get PVC for member")
		return errors.WithStack(err)
	}
	if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) {
		dbserverDataWillBeGone = true
	}

	// Once decided to drain the member, never go back
	if memberStatus.Phase == api.MemberPhaseDrain {
		dbserverDataWillBeGone = true
	}

	// Inspect cleaned out state
	c, err := r.context.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		log.Err(err).Debug("Failed to create member client")
		return errors.WithStack(err)
	}
	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		log.Err(err).Debug("Failed to access cluster")

		if r.context.GetSpec().Recovery.Get().GetAutoRecover() {
			if c, ok := k8sutil.GetContainerStatusByName(p, shared.ServerContainerName); ok {
				if t := c.State.Terminated; t != nil {
					return nil
				}
			}
		}
		return errors.WithStack(err)
	}
	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cleanedOut, err := cluster.IsCleanedOut(ctxChild, memberStatus.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	if cleanedOut {
		// Cleanout completed
		if memberStatus.Conditions.Update(api.ConditionTypeCleanedOut, true, "CleanedOut", "") {
			if err := updateMember(memberStatus); err != nil {
				return errors.WithStack(err)
			}
		}
		log.Debug("DBServer is cleaned out.")
		return nil
	}
	// Not cleaned out yet, check member status
	if memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		log.Warn("Member is already terminated before it could resign or be cleaned out. Not good, but removing dbserver pod because we cannot do anything further")
		// At this point we have to set CleanedOut to true,
		// because we can no longer reason about the state in the agency and
		// bringing back the dbserver again may result in an cleaned out server without us knowing
		if dbserverDataWillBeGone {
			memberStatus.Conditions.Update(api.ConditionTypeCleanedOut, true, "Draining server failed", "")
			memberStatus.CleanoutJobID = ""
			if memberStatus.Phase == api.MemberPhaseDrain {
				memberStatus.Phase = api.MemberPhaseCreated
			}
		} else if memberStatus.Phase == api.MemberPhaseResign {
			memberStatus.Phase = api.MemberPhaseCreated
		}

		if err := updateMember(memberStatus); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
	if memberStatus.Phase == api.MemberPhaseCreated {
		// No cleanout job triggered
		var jobID string
		ctxChild, cancelChild := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancelChild()

		ctxJobID := driver.WithJobIDResponse(ctxChild, &jobID)
		// Ensure the cleanout is triggered
		if dbserverDataWillBeGone {
			log.Debug("Server is not yet cleaned out. Triggering a clean out now")
			if err := cluster.CleanOutServer(ctxJobID, memberStatus.ID); err != nil {
				log.Err(err).Debug("Failed to clean out server")
				return errors.WithStack(err)
			}
			memberStatus.Phase = api.MemberPhaseDrain
		} else {
			log.Debug("Temporary shutdown, resign leadership")
			if err := cluster.ResignServer(ctxJobID, memberStatus.ID); err != nil {
				log.Err(err).Debug("Failed to resign server")
				return errors.WithStack(err)
			}
			memberStatus.Phase = api.MemberPhaseResign
		}

		memberStatus.CleanoutJobID = jobID

		if err := updateMember(memberStatus); err != nil {
			return errors.WithStack(err)
		}
	} else if memberStatus.Phase == api.MemberPhaseDrain {
		// Check the job progress
		cache, ok := r.context.GetAgencyCache()
		if !ok {
			return errors.Newf("AgencyCache is not ready")
		}

		details, jobStatus := cache.Target.GetJob(state.JobID(memberStatus.CleanoutJobID))
		switch jobStatus {
		case state.JobPhaseFailed:
			log.Str("reason", details.Reason).Warn("Job failed")
			// Revert cleanout state
			memberStatus.Phase = api.MemberPhaseCreated
			memberStatus.CleanoutJobID = ""
			if err := updateMember(memberStatus); err != nil {
				return errors.WithStack(err)
			}
			log.Error("Cleanout/Resign server job failed, continue anyway")
			return nil
		case state.JobPhaseFinished:
			memberStatus.CleanoutJobID = ""
			memberStatus.Phase = api.MemberPhaseCreated
		}
	} else if memberStatus.Phase == api.MemberPhaseResign {
		// Check the job progress
		cache, ok := r.context.GetAgencyCache()
		if !ok {
			return errors.Newf("AgencyCache is not ready")
		}

		details, jobStatus := cache.Target.GetJob(state.JobID(memberStatus.CleanoutJobID))
		switch jobStatus {
		case state.JobPhaseFailed:
			log.Str("reason", details.Reason).Warn("Resign Job failed")
			// Revert cleanout state
			memberStatus.Phase = api.MemberPhaseCreated
			memberStatus.CleanoutJobID = ""
			if err := updateMember(memberStatus); err != nil {
				return errors.WithStack(err)
			}
			log.Error("Cleanout/Resign server job failed, continue anyway")
			return nil
		case state.JobPhaseFinished:
			log.Str("reason", details.Reason).Debug("Resign Job finished")
			memberStatus.CleanoutJobID = ""
			memberStatus.Phase = api.MemberPhaseCreated
			if err := updateMember(memberStatus); err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
	}

	return errors.WithStack(errors.Newf("Server is not yet cleaned out"))

}
