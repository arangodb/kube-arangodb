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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func newCleanMemberServiceAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionCleanMemberService{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionCleanMemberService struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionCleanMemberService) Start(ctx context.Context) (bool, error) {
	member, exists := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !exists {
		a.log.Warn("Member does not exist")
		return true, nil
	}

	c := a.actionCtx.ACS().CurrentClusterCache()

	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	name := member.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)

	amember, ok := c.ArangoMember().V1().GetSimple(name)
	if !ok {
		a.log.Warn("ArangoMember does not exist")
		return true, nil
	}

	svc, ok := c.Service().V1().GetSimple(name)
	if !ok {
		a.log.Warn("Service does not exist")
		return true, nil
	}

	if !k8sutil.IsServiceRotationRequired(a.actionCtx.GetSpec(), svc) {
		return true, nil
	}

	if err := c.Client().Kubernetes().CoreV1().Services(c.Namespace()).Delete(ctxChild, amember.GetName(), meta.DeleteOptions{}); err != nil {
		a.log.Err(err).Warn("Unable to remove service")
		if !kerrors.IsNotFound(err) {
			return false, err
		}
	}

	return true, nil
}
