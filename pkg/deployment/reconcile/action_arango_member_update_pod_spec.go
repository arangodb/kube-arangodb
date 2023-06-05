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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// newArangoMemberUpdatePodSpecAction creates a new Action that implements the given
// planned ArangoMemberUpdatePodSpec action.
func newArangoMemberUpdatePodSpecAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionArangoMemberUpdatePodSpec{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

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
	status := a.actionCtx.GetStatus()

	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		a.log.Error("No such member")
		return true, nil
	}

	name := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)

	cache := a.actionCtx.ACS().CurrentClusterCache()

	_, ok := cache.ArangoMember().V1().GetSimple(name)
	if !ok {
		err := errors.Newf("ArangoMember not found")
		a.log.Err(err).Error("ArangoMember not found")
		return false, err
	}

	endpoint, err := a.actionCtx.GenerateMemberEndpoint(a.action.Group, m)
	if err != nil {
		a.log.Err(err).Error("Unable to render endpoint")
		return false, err
	}

	if m.Endpoint == nil || *m.Endpoint != endpoint {
		// Update endpoint
		m.Endpoint = util.NewType[string](endpoint)
		if err := status.Members.Update(m, a.action.Group); err != nil {
			a.log.Err(err).Error("Unable to update endpoint")
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

	renderedPod, err := a.actionCtx.RenderPodTemplateForMember(ctx, a.actionCtx.ACS(), spec, status, a.action.MemberID, imageInfo)
	if err != nil {
		a.log.Err(err).Error("Error while rendering pod")
		return false, err
	}

	checksum, err := resources.ChecksumArangoPod(groupSpec, resources.CreatePodFromTemplate(renderedPod))
	if err != nil {
		a.log.Err(err).Error("Error while getting pod checksum")
		return false, err
	}

	template, err := api.GetArangoMemberPodTemplate(renderedPod, checksum)
	if err != nil {
		a.log.Err(err).Error("Error while getting pod template")
		return false, err
	}

	if z := m.Endpoint; z != nil {
		q := *z
		template.Endpoint = &q
	}

	if err := inspector.WithArangoMemberUpdate(ctx, cache, name, func(member *api.ArangoMember) (bool, error) {
		if !member.Spec.Template.Equals(template) {
			member.Spec.Template = template.DeepCopy()
			return true, nil
		}

		return false, nil
	}); err != nil {
		a.log.Err(err).Error("Error while updating member")
		return false, err
	}

	if err := inspector.WithArangoMemberStatusUpdate(ctx, cache, name, func(member *api.ArangoMember) (bool, error) {
		if (member.Status.Template == nil || member.Status.Template.PodSpec == nil) && (m.Pod == nil || m.Pod.SpecVersion == "" || m.Pod.SpecVersion == template.PodSpecChecksum) {
			member.Status.Template = template.DeepCopy()
		}

		return true, nil
	}); err != nil {
		a.log.Err(err).Error("Error while updating member status")
		return false, err
	}

	return true, nil
}
