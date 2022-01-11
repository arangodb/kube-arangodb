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

package validation

import (
	"regexp"
	"strings"
)

var (
	dnsNameRegexp = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)
)

// IsValidDNSName returns true when the given input is a valid DNS name
func IsValidDNSName(input string) bool {
	// IsDNSName will validate the given string as a DNS name
	if input == "" || len(strings.Replace(input, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return dnsNameRegexp.MatchString(input)
}
