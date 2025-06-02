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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (x *AuthorizationV1PermissionRequest) Validate() error {
	if x == nil {
		x = &AuthorizationV1PermissionRequest{}
	}

	return shared.WithErrors(
		shared.PrefixResourceErrorFunc("user", func() error {
			return validateActionName(x.User)
		}),
		shared.PrefixResourceError("roles", shared.ValidateList(x.Roles, func(s string) error {
			return validateActionName(s)
		})),
		shared.PrefixResourceErrorFunc("action", func() error {
			return validateActionName(x.Action)
		}),
		shared.PrefixResourceErrorFunc("resource", func() error {
			return shared.ValidateRequiredNotEmpty(&x.Resource)
		}),
	)
}

func (x *AuthorizationV1PermissionRequest) Hash() string {
	if x == nil {
		return ""
	}
	return util.SHA256FromStringMap(map[string]string{
		"user":     util.SHA256FromString(x.GetUser()),
		"roles":    util.SHA256FromStringArray(x.GetRoles()...),
		"action":   util.SHA256FromString(x.GetAction()),
		"resource": util.SHA256FromString(x.GetResource()),
	})
}
