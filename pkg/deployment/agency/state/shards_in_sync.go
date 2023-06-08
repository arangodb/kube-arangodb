//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package state

import "time"

type ShardsSyncStatus map[string]time.Time

// NotInSyncSince returns a list of shards that have not been in sync for at least t.
func (s ShardsSyncStatus) NotInSyncSince(t time.Duration) []string {
	r := make([]string, 0, len(s))

	for k, v := range s {
		if v.IsZero() {
			continue
		}

		if time.Since(v) > t {
			r = append(r, k)
		}
	}

	return r
}
