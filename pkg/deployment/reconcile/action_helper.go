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

	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var (
	logger = logging.Global().RegisterAndGetLogger("action", logging.Info)
)

type actionEmpty struct {
	actionImpl
	actionEmptyStart
	actionEmptyCheckProgress
}

type actionEmptyCheckProgress struct {
}

// CheckProgress define optional check progress for action
// Returns: ready, abort, error.
func (e actionEmptyCheckProgress) CheckProgress(_ context.Context) (bool, bool, error) {
	return true, false, nil
}

type actionEmptyStart struct {
}

func (e actionEmptyStart) Start(_ context.Context) (bool, error) {
	return false, nil
}

func newActionImplDefRef(action api.Action, actionCtx ActionContext) actionImpl {
	return newActionImpl(action, actionCtx, &action.MemberID)
}

func newActionImpl(action api.Action, actionCtx ActionContext, memberIDRef *string) actionImpl {
	if memberIDRef == nil {
		panic("Action cannot have nil reference to member!")
	}

	return newBaseActionImpl(action, actionCtx, memberIDRef)
}

func newBaseActionImplDefRef(action api.Action, actionCtx ActionContext) actionImpl {
	return newBaseActionImpl(action, actionCtx, &action.MemberID)
}

func newBaseActionImpl(action api.Action, actionCtx ActionContext, memberIDRef *string) actionImpl {
	if memberIDRef == nil {
		panic("Action cannot have nil reference to member!")
	}

	a := actionImpl{
		action:      action,
		actionCtx:   actionCtx,
		memberIDRef: memberIDRef,
	}

	a.log = logger.Wrap(a.wrap)

	return a
}

type actionImpl struct {
	log       logging.Logger
	action    api.Action
	actionCtx ActionContext

	memberIDRef *string
}

func (a actionImpl) wrap(in *zerolog.Event) *zerolog.Event {
	in = in.
		Str("action-id", a.action.ID).
		Str("action-type", string(a.action.Type)).
		Str("group", a.action.Group.AsRole()).
		Str("member-id", a.action.MemberID)

	if status := a.actionCtx.GetStatus(); status.Members.ContainsID(a.action.MemberID) {
		if member, _, ok := status.Members.ElementByID(a.action.MemberID); ok {
			in = in.Str("phase", string(member.Phase))
		}
	}

	for k, v := range a.action.Params {
		in = in.Str("param."+k, v)
	}

	for k, v := range a.action.Locals {
		in = in.Str("local."+k.String(), v)
	}

	return in
}

// MemberID returns the member ID used / created in the current action.
func (a actionImpl) MemberID() string {
	return *a.memberIDRef
}
