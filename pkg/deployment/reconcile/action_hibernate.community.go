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
//go:build !enterprise
// +build !enterprise

package reconcile

import (
	"context"
)

// actionHibernate implements a hibernation action.
type actionHibernate struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start does nothing in community version.
func (a *actionHibernate) Start(_ context.Context) (bool, error) {
	return true, nil
}

// CheckProgress does nothing in community version.
func (a *actionHibernate) CheckProgress(_ context.Context) (bool, bool, error) {
	return true, false, nil
}
