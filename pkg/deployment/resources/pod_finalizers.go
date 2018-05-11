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

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// runPodFinalizers goes through the list of pod finalizers to see if they can be removed.
func (r *Resources) runPodFinalizers(ctx context.Context, p *v1.Pod, memberStatus api.MemberStatus) error {
	log := r.log.With().Str("pod-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerDrainDBServer:
			log.Debug().Msg("Inspecting drain dbserver finalizer")
			if err := r.inspectFinalizerDrainDBServer(ctx, log, p, memberStatus); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		kubecli := r.context.GetKubeCli()
		if err := k8sutil.RemovePodFinalizers(log, kubecli, p, removalList); err != nil {
			log.Debug().Err(err).Msg("Failed to update pod (to remove finalizers)")
			return maskAny(err)
		}
	}
	return nil
}

// inspectFinalizerDrainDBServer checks the finalizer condition for drain-dbserver.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerDrainDBServer(ctx context.Context, log zerolog.Logger, p *v1.Pod, memberStatus api.MemberStatus) error {
	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug().Msg("Pod is already failed, safe to remove drain dbserver finalizer")
		return nil
	}
	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug().Msg("Entire deployment is being deleted, safe to remove drain dbserver finalizer")
		return nil
	}
	// Inspect cleaned out state
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
		// All done
		log.Debug().Msg("Server is cleaned out. Save to remove drain dbserver finalizer")
		return nil
	}
	// Not cleaned out yet, check member status
	if memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		log.Warn().Msg("Member is already terminated before it could be cleaned out. Not good, but removing drain dbserver finalizer because we cannot do anything further")
		return nil
	}
	// Ensure the cleanout is triggered
	log.Debug().Msg("Server is not yet clean out. Triggering a clean out now")
	if err := cluster.CleanOutServer(ctx, memberStatus.ID); err != nil {
		log.Debug().Err(err).Msg("Failed to clean out server")
		return maskAny(err)
	}
	return maskAny(fmt.Errorf("Server is not yet cleaned out"))
}
