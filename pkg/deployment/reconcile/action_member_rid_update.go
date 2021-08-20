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

	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeMemberRIDUpdate, newMemberRIDUpdate)
}

func newMemberRIDUpdate(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &memberRIDUpdateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type memberRIDUpdateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *memberRIDUpdateAction) Start(ctx context.Context) (bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	m.RID = uuid.NewUUID()

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}
