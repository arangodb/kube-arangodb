//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
)

func newMemberStatusSyncAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionMemberStatusSync{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionMemberStatusSync struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionMemberStatusSync) ReloadComponents() []definitions.Component {
	return []definitions.Component{
		definitions.ArangoMember,
	}
}

func (a *actionMemberStatusSync) Start(ctx context.Context) (bool, error) {
	m, g, ok := a.actionCtx.GetMemberStatusAndGroupByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	if g != a.action.Group {
		a.log.Error("Invalid group")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		a.log.Error("Unable to get cache")
		return true, nil
	}

	name := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)

	amember, ok := cache.ArangoMember().V1().GetSimple(name)
	if !ok {
		a.log.Error("Unable to find ArangoMember")
		return true, nil
	}

	amemberc := amember.DeepCopy()
	if amemberc.Status.Propagate(m) {
		// Change applied
		nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer c()

		if _, err := cache.ArangoMemberModInterface().V1().UpdateStatus(nctx, amemberc, meta.UpdateOptions{}); err != nil {
			return false, errors.WithStack(err)
		}

		if err := cache.Refresh(nctx); err != nil {
			return false, errors.WithStack(err)
		}
	}

	return true, nil
}
