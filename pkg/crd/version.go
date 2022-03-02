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

package crd

import (
	"strconv"
	"strings"

	"github.com/arangodb/go-driver"
)

func isVersionValid(a driver.Version) bool {
	q := strings.SplitN(string(a), ".", 3)

	if len(q) < 2 {
		// We do not have 2 parts
		return false
	}

	_, err := strconv.Atoi(q[0])
	if err != nil {
		return false
	}

	_, err = strconv.Atoi(q[1])
	return err == nil
}

func isUpdateRequired(a, b driver.Version) bool {
	if a == b {
		return false
	}

	if !isVersionValid(b) {
		return true
	}

	return a.CompareTo(b) > 0
}
