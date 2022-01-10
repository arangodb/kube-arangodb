//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package options

import "strings"

var (
	criticalOptionKeys = map[string]struct{}{
		"cluster.jwt-secret":   struct{}{},
		"cluster.endpoint":     struct{}{},
		"master.endpoint":      struct{}{},
		"master.jwt-secret":    struct{}{},
		"mq.type":              struct{}{},
		"server.client-cafile": struct{}{},
		"server.endpoint":      struct{}{},
		"server.keyfile":       struct{}{},
		"server.port":          struct{}{},
	}
)

// IsCriticalOption returns true if the given string is the key of
// an option of arangosync that cannot be overwritten.
func IsCriticalOption(optionKey string) bool {
	optionKey = strings.TrimPrefix(optionKey, "--")
	_, found := criticalOptionKeys[optionKey]
	return found
}
