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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

type Effect string

func (a Effect) Validate() error {
	switch a {
	case EffectAllow, EffectDeny:
		return nil
	}

	return errors.Errorf("Invalid effect `%s`: Accepted only Allow or Deny", a)
}

const (
	EffectAllow Effect = "Allow"
	EffectDeny  Effect = "Deny"
)
