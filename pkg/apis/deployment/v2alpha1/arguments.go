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

package v2alpha1

import (
	"strings"

	arangodOptions "github.com/arangodb/kube-arangodb/pkg/util/arangod/options"
	arangosyncOptions "github.com/arangodb/kube-arangodb/pkg/util/arangosync/options"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Arguments []string

func (a Arguments) Validate(group ServerGroup) error {
	for _, arg := range a {
		parts := strings.Split(arg, "=")
		optionKey := strings.TrimSpace(parts[0])
		switch group.Type() {
		case ServerGroupTypeArangoD:
			if arangodOptions.IsCriticalOption(optionKey) {
				return errors.WithStack(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
			}
		case ServerGroupTypeArangoSync:
			if arangosyncOptions.IsCriticalOption(optionKey) {
				return errors.WithStack(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
			}
		}
	}

	return nil
}
