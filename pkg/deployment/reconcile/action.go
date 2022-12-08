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
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
)

const (
	DefaultStartFailureGracePeriod = 10 * time.Second
)

func GetAllActions() []api.ActionType {
	z := make([]api.ActionType, 0, len(definedActions))

	for k := range definedActions {
		z = append(z, k)
	}

	return z
}

// ActionCore executes a single Plan item.
type ActionCore interface {
	// Start performs the start of the action.
	// Returns true if the action is completely finished, false in case
	// the start time needs to be recorded and a ready condition needs to be checked.
	Start(ctx context.Context) (bool, error)
	// CheckProgress checks the progress of the action.
	// Returns: ready, abort, error.
	CheckProgress(ctx context.Context) (bool, bool, error)
}

// Action executes a single Plan item.
type Action interface {
	ActionCore

	// MemberID Return the MemberID used / created in this action
	MemberID() string
}

// ActionPost keep interface which is executed after action is completed.
type ActionPost interface {
	Action

	// Post execute after action is completed
	Post(ctx context.Context) error
}

func getActionPost(a Action, ctx context.Context) error {
	if c, ok := a.(ActionPost); !ok {
		return nil
	} else {
		return c.Post(ctx)
	}
}

// ActionPre keep interface which is executed before action is started.
type ActionPre interface {
	Action

	// Pre execute after action is completed
	Pre(ctx context.Context) error
}

func getActionPre(a Action, ctx context.Context) error {
	if c, ok := a.(ActionPre); !ok {
		return nil
	} else {
		return c.Pre(ctx)
	}
}

// ActionReloadCachedStatus keeps information about CachedStatus reloading (executed after action has been executed)
type ActionReloadCachedStatus interface {
	Action

	// ReloadComponents return cache components to be reloaded
	ReloadComponents() (types.UID, []definitions.Component)
}

func getActionReloadCachedStatus(a Action) (types.UID, []definitions.Component) {
	if c, ok := a.(ActionReloadCachedStatus); !ok {
		return "", nil
	} else {
		return c.ReloadComponents()
	}
}

// ActionStartFailureGracePeriod extend action definition to allow specifying start failure grace period
type ActionStartFailureGracePeriod interface {
	Action

	// StartFailureGracePeriod returns information about failure grace period (defaults to 0)
	StartFailureGracePeriod() time.Duration
}

func wrapActionStartFailureGracePeriod(action Action, failureGracePeriod time.Duration) Action {
	return &actionStartFailureGracePeriod{
		Action:             action,
		failureGracePeriod: failureGracePeriod,
	}
}

func withActionStartFailureGracePeriod(in actionFactory, failureGracePeriod time.Duration) actionFactory {
	return func(action api.Action, actionCtx ActionContext) Action {
		return wrapActionStartFailureGracePeriod(in(action, actionCtx), failureGracePeriod)
	}
}

var _ ActionStartFailureGracePeriod = &actionStartFailureGracePeriod{}

type actionStartFailureGracePeriod struct {
	Action
	failureGracePeriod time.Duration
}

func (a actionStartFailureGracePeriod) StartFailureGracePeriod() time.Duration {
	return a.failureGracePeriod
}

func getStartFailureGracePeriod(a Action) time.Duration {
	if c, ok := a.(ActionStartFailureGracePeriod); !ok {
		return DefaultStartFailureGracePeriod
	} else {
		return c.StartFailureGracePeriod()
	}
}

// ActionPlanAppender modify plan after action execution
type ActionPlanAppender interface {
	Action

	// ActionPlanAppender modify plan after action execution
	ActionPlanAppender(current api.Plan) (api.Plan, bool)
}

func getActionPlanAppender(a Action, plan api.Plan) (api.Plan, bool) {
	if c, ok := a.(ActionPlanAppender); !ok {
		return plan, false
	} else {
		return c.ActionPlanAppender(plan)
	}
}

type actionFactory func(action api.Action, actionCtx ActionContext) Action

var (
	definedActions     = map[api.ActionType]actionFactory{}
	definedActionsLock sync.Mutex
)

func registerAction(t api.ActionType, f actionFactory) {
	definedActionsLock.Lock()
	defer definedActionsLock.Unlock()

	_, ok := definedActions[t]
	if ok {
		panic(fmt.Sprintf("Action already defined %s", t))
	}

	definedActions[t] = f
}

func getActionFactory(t api.ActionType) (actionFactory, bool) {
	definedActionsLock.Lock()
	defer definedActionsLock.Unlock()

	f, ok := definedActions[t]
	return f, ok
}

type actionSuccess struct{}

// NewActionSuccess returns action which always returns success.
func NewActionSuccess() ActionCore {
	return actionSuccess{}
}

// Start always returns false to start with progress.
func (actionSuccess) Start(_ context.Context) (bool, error) {
	return false, nil
}

// CheckProgress always returns true.
func (actionSuccess) CheckProgress(_ context.Context) (bool, bool, error) {
	return true, false, nil
}
