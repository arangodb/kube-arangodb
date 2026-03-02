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

package types

import (
	"sort"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func (x *Policy) Hash() string {
	if x == nil {
		return ""
	}

	return util.SHA256FromStringArray(
		util.SHA256FromHashArray(x.GetStatements()),
	)
}

func (x *Policy) Deleted() bool {
	return x == nil
}

func (x *Policy) Clean() error {
	if x == nil {
		return nil
	}

	for _, stmt := range x.Statements {
		if err := stmt.Clean(); err != nil {
			return err
		}
	}

	return nil
}

func (x *Policy) Validate() error {
	if x == nil {
		return nil
	}

	return errors.Errors(
		shared.PrefixResourceError("statements", shared.ValidateInterfaceList(x.GetStatements())),
	)
}

func (x *PolicyStatement) Hash() string {
	if x == nil {
		return ""
	}

	return util.SHA256FromStringArray(
		util.SHA256FromString(x.GetEffect().String()),
		util.SHA256FromStringArray(x.GetActions()...),
		util.SHA256FromStringArray(x.GetResources()...),
	)
}

func (x *PolicyStatement) Clean() error {
	if x == nil {
		return nil
	}

	sort.Strings(x.Actions)
	sort.Strings(x.Resources)

	x.Actions = util.UniqueList(x.Actions)
	x.Resources = util.UniqueList(x.Resources)

	return nil
}

func (x *PolicyStatement) Validate() error {
	if x == nil {
		return nil
	}

	return errors.Errors(
		shared.ValidateOptionalInterfacePath("effect", x.GetEffect()),
		shared.PrefixResourceError("actions", shared.ValidateList(x.GetActions(), ValidateAction)),
		shared.PrefixResourceError("resources", shared.ValidateList(x.GetResources(), ValidateResource)),
	)
}

func (x *PolicyStatement) Match(action, resource string, context map[string][]string) bool {
	if x == nil {
		return false
	}

	return false
}

func (x Effect) Validate() error {
	switch x {
	case Effect_Allow, Effect_Deny:
		return nil
	}

	return errors.Errorf("Unknown Effect: %s", x)
}

func ValidateAction(action string) error {
	if z := strings.Split(action, ":"); len(z) != 2 {
		return errors.Errorf("Invalid action '%s': expected format '<namespace>:<name>'", action)
	} else {
		if z[0] == "" {
			return errors.Errorf("Invalid action '%s': empty namespace", action)
		}

		if z[1] == "" {
			return errors.Errorf("Invalid action '%s': empty name", action)
		}
	}

	return nil
}

func ValidateResource(resource string) error {
	if resource == "" {
		return errors.Errorf("Invalid resource '%s': empty name", resource)
	}
	return nil
}
