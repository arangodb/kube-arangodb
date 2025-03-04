//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package definition

import (
	"regexp"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	actionNameRE = regexp.MustCompile(`^[a-z]+(\.[a-z]+)*$`)
)

func validateActionName(name string) error {
	if !actionNameRE.MatchString(name) {
		return errors.Errorf("Action `%s` does not match the regex", name)
	}

	return nil
}

var _ shared.ValidateInterface = &AuthorizationV1Policy{}

func (x *AuthorizationV1Action) Validate() error {
	if x == nil {
		return errors.Errorf("Actions is not allowed to be nil")
	}

	return shared.WithErrors(
		shared.PrefixResourceError("action", validateActionName(x.Name)),
		shared.ValidateRequiredNotEmptyPath("description", &x.Description),
		shared.PrefixResourceError("subActions", shared.ValidateList(x.SubActions, func(s string) error {
			return validateActionName(s)
		})),
	)
}

func (x AuthorizationV1Effect) Validate() error {
	switch x {
	case AuthorizationV1Effect_Allow, AuthorizationV1Effect_Deny:
		return nil
	}

	return errors.Errorf("Invalid Effect value")
}

func (x *AuthorizationV1Statement) Validate() error {
	if x == nil {
		return errors.Errorf("Statement is not allowed to be nil")
	}

	return shared.WithErrors(
		shared.ValidateRequiredInterfacePath("effect", x.Effect),
		shared.ValidateRequiredNotEmptyPath("description", &x.Description),
		shared.PrefixResourceError("actions", shared.ValidateList(x.Actions, func(s string) error {
			return validateActionName(s)
		})),
		shared.PrefixResourceError("resources", shared.ValidateList(x.Resources, func(s string) error {
			return shared.ValidateRequiredNotEmpty(&s)
		})),
	)
}

func (x *AuthorizationV1Policy) Validate() error {
	if x == nil {
		return errors.Errorf("Statement is not allowed to be nil")
	}

	return shared.WithErrors(
		shared.PrefixResourceError("name", validateActionName(x.Name)),
		shared.ValidateRequiredNotEmptyPath("description", &x.Description),
		shared.PrefixResourceError("statements", shared.ValidateInterfaceList(x.Statements)),
	)
}
