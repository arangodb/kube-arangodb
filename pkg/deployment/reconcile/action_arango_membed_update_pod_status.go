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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeArangoMemberUpdatePodStatus, newArangoMemberUpdatePodStatusAction)
}

// newArangoMemberUpdatePodStatusAction creates a new Action that implements the given
// planned ArangoMemberUpdatePodStatus action.
func newArangoMemberUpdatePodStatusAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionArangoMemberUpdatePodStatus{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, rotateMemberTimeout)

	return a
}

// actionArangoMemberUpdatePodStatus implements an ArangoMemberUpdatePodStatus.
type actionArangoMemberUpdatePodStatus struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionArangoMemberUpdatePodStatus) Start(ctx context.Context) (bool, error) {
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

	if member.Status.TemplateChecksum != member.Spec.TemplateChecksum || member.Status.Template == nil {
		if err := a.actionCtx.WithArangoMemberStatusUpdate(context.Background(), member.GetNamespace(), member.GetName(), func(obj *api.ArangoMember, status *api.ArangoMemberStatus) bool {
			if status.TemplateChecksum != obj.Spec.TemplateChecksum || status.Template == nil {
				status.TemplateChecksum = obj.Spec.TemplateChecksum
				status.Template = obj.Spec.Template.DeepCopy()
				return true
			}
			return false
		}); err != nil {
			log.Err(err).Msg("Error while updating member")
			return false, err
		}
	}

	return false, nil
}
