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
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/util"
	core "k8s.io/api/core/v1"
)

func secretKeysToListWithPrefix(s *core.Secret) []string {
	return util.PrefixStringArray(secretKeysToList(s), "sha256:")
}

func secretKeysToList(s *core.Secret) []string {
	keys := make([]string, 0, len(s.Data))

	for key := range s.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
