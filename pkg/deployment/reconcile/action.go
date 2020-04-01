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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"
	"fmt"
	"sync"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// Action executes a single Plan item.
type Action interface {
	// Start performs the start of the action.
	// Returns true if the action is completely finished, false in case
	// the start time needs to be recorded and a ready condition needs to be checked.
	Start(ctx context.Context) (bool, error)
	// CheckProgress checks the progress of the action.
	// Returns: ready, abort, error.
	CheckProgress(ctx context.Context) (bool, bool, error)
	// Timeout returns the amount of time after which this action will timeout.
	Timeout() time.Duration
	// Return the MemberID used / created in this action
	MemberID() string
}

type actionFactory func(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action

var (
	actions     = map[api.ActionType]actionFactory{}
	actionsLock sync.Mutex
)

func registerAction(t api.ActionType, f actionFactory) {
	actionsLock.Lock()
	defer actionsLock.Unlock()

	_, ok := actions[t]
	if ok {
		panic(fmt.Sprintf("Action already defined %s", t))
	}

	actions[t] = f
}

func getActionFactory(t api.ActionType) (actionFactory, bool) {
	actionsLock.Lock()
	defer actionsLock.Unlock()

	f, ok := actions[t]
	return f, ok
}
