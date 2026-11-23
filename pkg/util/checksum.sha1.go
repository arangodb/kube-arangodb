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

package util

import (
	"crypto/sha1"
	"fmt"
)

// SHA1FromString returns the hex SHA-1 checksum of data.
//
// SHA-1 is not FIPS-approved and panics under GODEBUG=fips140=only. Use it only through the
// FIPS-gated ShortHash helper, which selects SHA-256 when running in FIPS mode.
func SHA1FromString(data string) string {
	return SHA1([]byte(data))
}

// SHA1 returns the hex SHA-1 checksum of data.
//
// SHA-1 is not FIPS-approved and panics under GODEBUG=fips140=only. Use it only through the
// FIPS-gated ShortHash helper, which selects SHA-256 when running in FIPS mode.
func SHA1(data []byte) string {
	return fmt.Sprintf("%0x", sha1.Sum(data))
}
