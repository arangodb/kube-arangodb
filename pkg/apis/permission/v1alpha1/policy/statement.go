//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package policy

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Statements []Statement

func (a Statements) Validate() error {
	return shared.ValidateInterfaceList(a)
}

type Statement struct {
	// Effect defines the statement effect.
	// +doc/enum: Allow|Action is Allowed
	// +doc/enum: Deny|Action is Denied
	// +doc/required
	Effect Effect `json:"effect"`

	// Actions defines the list of actions.
	// Action needs to be defined in format `<namespace>:<name>`
	// +doc/required
	Actions Actions `json:"actions"`

	// Resources defines the list of resources
	// +doc/required
	Resources Resources `json:"resources"`
}

func (a Statement) Validate() error {
	return errors.Errors(
		shared.ValidateRequiredInterfacePath("effect", a.Effect),
		shared.ValidateRequiredInterfacePath("actions", a.Actions),
		shared.ValidateRequiredInterfacePath("resources", a.Resources),
	)
}
