//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeArangoMemberUpdatePodSpec, newArangoMemberUpdatePodSpecAction)
}

// newArangoMemberUpdatePodSpecAction creates a new Action that implements the given
// planned ArangoMemberUpdatePodSpec action.
func newArangoMemberUpdatePodSpecAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionArangoMemberUpdatePodSpec{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

var _ ActionReloadCachedStatus = &actionArangoMemberUpdatePodSpec{}

// actionArangoMemberUpdatePodSpec implements an ArangoMemberUpdatePodSpec.
type actionArangoMemberUpdatePodSpec struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionArangoMemberUpdatePodSpec) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()
	status := a.actionCtx.GetStatusSnapshot()

	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, nil
	}

	cache := a.actionCtx.GetCachedStatus()

	member, ok := cache.ArangoMember(m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group))
	if !ok {
		err := errors.Newf("ArangoMember not found")
		log.Error().Err(err).Msg("ArangoMember not found")
		return false, err
	}

	endpoint, err := pod.GenerateMemberEndpoint(a.actionCtx.GetCachedStatus(), a.actionCtx.GetAPIObject(), spec, a.action.Group, m)
	if err != nil {
		log.Error().Err(err).Msg("Unable to render endpoint")
		return false, err
	}

	if m.Endpoint == nil || *m.Endpoint != endpoint {
		// Update endpoint
		m.Endpoint = &endpoint
		if err := status.Members.Update(m, a.action.Group); err != nil {
			log.Error().Err(err).Msg("Unable to update endpoint")
			return false, err
		}
	}

	groupSpec := spec.GetServerGroupSpec(a.action.Group)

	imageInfo, imageFound := a.actionCtx.SelectImage(spec, status)
	if !imageFound {
		// Image is not found, so rotation is not needed
		return true, nil
	}

	if m.Image != nil {
		imageInfo = *m.Image
	}

	renderedPod, err := a.actionCtx.RenderPodTemplateForMember(ctx, a.actionCtx.GetCachedStatus(), spec, status, a.action.MemberID, imageInfo)
	if err != nil {
		log.Err(err).Msg("Error while rendering pod")
		return false, err
	}

	checksum, err := resources.ChecksumArangoPod(groupSpec, resources.CreatePodFromTemplate(renderedPod))
	if err != nil {
		log.Err(err).Msg("Error while getting pod checksum")
		return false, err
	}

	template, err := api.GetArangoMemberPodTemplate(renderedPod, checksum)
	if err != nil {
		log.Err(err).Msg("Error while getting pod template")
		return false, err
	}

	if err := a.actionCtx.WithArangoMemberUpdate(context.Background(), member.GetNamespace(), member.GetName(), func(member *api.ArangoMember) bool {
		if !member.Spec.Template.Equals(template) {
			member.Spec.Template = template.DeepCopy()
			return true
		}

		return false
	}); err != nil {
		log.Err(err).Msg("Error while updating member")
		return false, err
	}

	if err := a.actionCtx.WithArangoMemberStatusUpdate(context.Background(), member.GetNamespace(), member.GetName(), func(member *api.ArangoMember, status *api.ArangoMemberStatus) bool {
		if (status.Template == nil || status.Template.PodSpec == nil) && (m.PodSpecVersion == "" || m.PodSpecVersion == template.PodSpecChecksum) {
			status.Template = template.DeepCopy()
		}

		return true
	}); err != nil {
		log.Err(err).Msg("Error while updating member status")
		return false, err
	}

	return true, nil
}

func (a *actionArangoMemberUpdatePodSpec) ReloadCachedStatus() bool {
	return true
}
