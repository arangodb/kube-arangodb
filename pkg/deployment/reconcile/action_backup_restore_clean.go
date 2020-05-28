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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeBackupRestoreClean, newBackupRestoreCleanAction)
}

func newBackupRestoreCleanAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionBackupRestoreClean{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, backupRestoreTimeout)

	return a
}

// actionBackupRestoreClean implements an BackupRestoreClean.
type actionBackupRestoreClean struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a actionBackupRestoreClean) Start(ctx context.Context) (bool, error) {
	if err := a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
		if s.Restore == nil {
			return false
		}

		s.Restore = nil
		return true
	}, true); err != nil {
		return false, err
	}

	return true, nil
}
