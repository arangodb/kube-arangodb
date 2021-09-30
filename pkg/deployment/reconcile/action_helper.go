//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

type TimeoutFetcher func(deploymentSpec api.DeploymentSpec) time.Duration

func NewTimeoutFetcher(t time.Duration) TimeoutFetcher {
	return func(deploymentSpec api.DeploymentSpec) time.Duration {
		return t
	}
}

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

func newActionImplDefRef(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout time.Duration) actionImpl {
	return newActionImpl(log, action, actionCtx, timeout, &action.MemberID)
}

func newActionImpl(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout time.Duration, memberIDRef *string) actionImpl {
	if memberIDRef == nil {
		panic("Action cannot have nil reference to member!")
	}

	return newBaseActionImpl(log, action, actionCtx, NewTimeoutFetcher(timeout), memberIDRef)
}

func newBaseActionImplDefRef(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout TimeoutFetcher) actionImpl {
	return newBaseActionImpl(log, action, actionCtx, timeout, &action.MemberID)
}

func newBaseActionImpl(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout TimeoutFetcher, memberIDRef *string) actionImpl {
	if memberIDRef == nil {
		panic("Action cannot have nil reference to member!")
	}

	return actionImpl{
		log:         log,
		action:      action,
		actionCtx:   actionCtx,
		timeout:     timeout,
		memberIDRef: memberIDRef,
	}
}

type actionImpl struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext

	timeout     TimeoutFetcher
	memberIDRef *string
}

// Timeout returns the amount of time after which this action will timeout.
func (a actionImpl) Timeout(deploymentSpec api.DeploymentSpec) time.Duration {
	if a.timeout == nil {
		return defaultTimeout
	}

	return a.timeout(deploymentSpec)
}

// MemberID returns the member ID used / created in the current action.
func (a actionImpl) MemberID() string {
	return *a.memberIDRef
}
