//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func getShutdownHelper(a *api.Action, ctx ActionContext, log zerolog.Logger) ActionCore {
	serverGroup := ctx.GetSpec().GetServerGroupSpec(a.Group)

	switch serverGroup.ShutdownMethod.Get() {
	case api.ServerGroupShutdownMethodDelete:
		return shutdownHelperDelete{action: a, actionCtx: ctx, log: log}
	default:
		return shutdownHelperAPI{action: a, actionCtx: ctx, log: log}
	}
}

type shutdownHelperAPI struct {
	log       zerolog.Logger
	action    *api.Action
	actionCtx ActionContext
}

func (s shutdownHelperAPI) Start(ctx context.Context) (bool, error) {
	log := s.log

	log.Info().Msgf("Using API to shutdown member")

	group := s.action.Group
	m, ok := s.actionCtx.GetMemberStatusByID(s.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}
	if m.PodName == "" {
		log.Warn().Msgf("Pod is empty")
		return true, nil
	}
	// Remove finalizers, so Kubernetes will quickly terminate the pod
	if err := s.actionCtx.RemovePodFinalizers(ctx, m.PodName); err != nil {
		return false, errors.WithStack(err)
	}
	if group.IsArangod() {
		// Invoke shutdown endpoint
		ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
		defer cancel()
		c, err := s.actionCtx.GetServerClient(ctxChild, group, s.action.MemberID)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create member client")
			return false, errors.WithStack(err)
		}
		removeFromCluster := false
		log.Debug().Bool("removeFromCluster", removeFromCluster).Msg("Shutting down member")
		ctxChild, cancel = context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := c.Shutdown(ctxChild, removeFromCluster); err != nil {
			// Shutdown failed. Let's check if we're already done
			if ready, _, err := s.CheckProgress(ctxChild); err == nil && ready {
				// We're done
				return true, nil
			}
			log.Debug().Err(err).Msg("Failed to shutdown member")
			return false, errors.WithStack(err)
		}
	} else if group.IsArangosync() {
		// Terminate pod
		if err := s.actionCtx.DeletePod(ctx, m.PodName); err != nil {
			return false, errors.WithStack(err)
		}
	}

	return false, nil
}

func (s shutdownHelperAPI) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	log := s.log
	m, found := s.actionCtx.GetMemberStatusByID(s.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, false, nil
	}
	if !m.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Pod is not yet terminated
		return false, false, nil
	}

	return true, false, nil
}

type shutdownHelperDelete struct {
	log       zerolog.Logger
	action    *api.Action
	actionCtx ActionContext
}

func (s shutdownHelperDelete) Start(ctx context.Context) (bool, error) {
	log := s.log

	log.Info().Msgf("Using Pod Delete to shutdown member")

	m, ok := s.actionCtx.GetMemberStatusByID(s.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	if m.PodName == "" {
		log.Warn().Msgf("Pod is empty")
		return true, nil
	}

	// Terminate pod
	if err := s.actionCtx.DeletePod(ctx, m.PodName); err != nil {
		if !k8sutil.IsNotFound(err) {
			return false, errors.WithStack(err)
		}

	}

	return false, nil
}

func (s shutdownHelperDelete) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	log := s.log
	m, found := s.actionCtx.GetMemberStatusByID(s.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, false, nil
	}

	if !m.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Pod is not yet terminated
		log.Warn().Msgf("Pod not yet terminated")
		return false, false, nil
	}

	if m.PodName != "" {
		if _, err := s.actionCtx.GetPod(ctx, m.PodName); err == nil {
			log.Warn().Msgf("Pod still exists")
			return false, false, nil
		} else if !k8sutil.IsNotFound(err) {
			return false, false, errors.WithStack(err)
		}
	}

	return true, false, nil
}
