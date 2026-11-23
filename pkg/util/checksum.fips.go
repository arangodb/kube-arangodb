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

import "crypto/fips140"

// Checksum returns a hex content checksum. In FIPS mode it uses SHA-224; otherwise MD5.
//
// Gating on the Go FIPS runtime mode keeps the historical MD5 CRD and CA checksums on non-FIPS
// operators, so an upgrade does not change them and trigger a spurious re-apply, while remaining
// FIPS-approved under GODEBUG=fips140=only, where MD5 is refused.
func Checksum(data []byte) string {
	if fips140.Enabled() {
		return SHA224(data)
	}
	return MD5(data)
}

// ChecksumFromString returns a hex content checksum of data. See Checksum.
func ChecksumFromString(data string) string {
	return Checksum([]byte(data))
}

// ShortHash returns the first 6 hex characters of a hash of data, used for pod-name and
// resource-name suffixes. In FIPS mode it uses SHA-256; otherwise SHA-1.
//
// Gating on the Go FIPS runtime mode keeps the historical SHA-1 suffixes on non-FIPS operators, so an
// upgrade does not rename existing pods and resources, while remaining FIPS-approved under
// GODEBUG=fips140=only, where SHA-1 is refused.
func ShortHash(data []byte) string {
	if fips140.Enabled() {
		return SHA256(data)[:6]
	}
	return SHA1(data)[:6]
}
