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
		"agency.activate":                struct{}{},
		"agency.disaster-recovery-id":    struct{}{},
		"agency.endpoint":                struct{}{},
		"agency.my-address":              struct{}{},
		"agency.size":                    struct{}{},
		"agency.supervision":             struct{}{},
		"cluster.agency-endpoint":        struct{}{},
		"cluster.my-address":             struct{}{},
		"cluster.my-role":                struct{}{},
		"database.directory":             struct{}{},
		"database.auto-upgrade":          struct{}{},
		"foxx.queues":                    struct{}{},
		"replication.automatic-failover": struct{}{},
		"rocksdb.encryption-keyfile":     struct{}{},
		"server.authentication":          struct{}{},
		"server.endpoint":                struct{}{},
		"server.jwt-secret":              struct{}{},
		"server.storage-engine":          struct{}{},
		"ssl.keyfile":                    struct{}{},
		"ssl.ecdh-curve":                 struct{}{},
	}
)

// IsCriticalOption returns true if the given string is the key of
// an option of arangod that cannot be overwritten.
func IsCriticalOption(optionKey string) bool {
	optionKey = strings.TrimPrefix(optionKey, "--")
	_, found := criticalOptionKeys[optionKey]
	return found
}
