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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

const (
	ActionTypeArangoMemberUpdatePodStatusChecksum = "checksum"
)

// newArangoMemberUpdatePodStatusAction creates a new Action that implements the given
// planned ArangoMemberUpdatePodStatus action.
func newArangoMemberUpdatePodStatusAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionArangoMemberUpdatePodStatus{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

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
		a.log.Error("No such member")
		return true, nil
	}

	name := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)
	cache := a.actionCtx.ACS().CurrentClusterCache()

	member, ok := cache.ArangoMember().V1().GetSimple(name)
	if !ok {
		err := errors.Newf("ArangoMember not found")
		a.log.Err(err).Error("ArangoMember not found")
		return false, err
	}

	if c, ok := a.action.GetParam(ActionTypeArangoMemberUpdatePodStatusChecksum); ok {
		if member.Spec.Template == nil {
			return true, nil
		}

		if member.Spec.Template.Checksum != c {
			// Checksum is invalid
			return true, nil
		}
	}

	if member.Status.Template == nil || !member.Status.Template.Equals(member.Spec.Template) {
		if err := inspector.WithArangoMemberStatusUpdate(ctx, cache, name, func(in *api.ArangoMember) (bool, error) {
			if in.Status.Template == nil || !in.Status.Template.Equals(member.Spec.Template) {
				in.Status.Template = member.Spec.Template.DeepCopy()
				return true, nil
			}
			return false, nil
		}); err != nil {
			a.log.Err(err).Error("Error while updating member")
			return false, err
		}
	}

	return true, nil
}
