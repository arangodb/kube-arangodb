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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (x *Role) Hash() string {
	if x == nil {
		return ""
	}

	return util.SHA256FromStringArray(
		util.SHA256FromStringArray(x.GetPolicies()...),
		x.GetScope().Hash(),
	)
}

func (x *Role) Deleted() bool {
	return x == nil
}

func (x *Role) Clean() error {
	if x == nil {
		return nil
	}

	sort.Strings(x.Policies)

	x.Policies = util.UniqueList(x.Policies)

	if err := x.GetScope().Clean(); err != nil {
		return err
	}

	return nil
}

func (x *Role) Validate() error {
	if x == nil {
		return nil
	}

	if x.Scope == nil {
		return errors.Errorf("scope is required")
	}

	return x.GetScope().Validate()
}
