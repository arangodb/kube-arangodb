//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package inspector

import (
	"fmt"

	"github.com/arangodb/go-driver"
)

func newMinK8SVersion(ver driver.Version) error {
	return minK8SVersion{ver: ver}
}

type minK8SVersion struct {
	ver driver.Version
}

func (m minK8SVersion) Error() string {
	return fmt.Sprintf("Kubernetes %s or lower is not supported anymore", m.ver)
}

func IsK8SVersion(err error) (driver.Version, bool) {
	if err == nil {
		return "", false
	}

	if v, ok := err.(minK8SVersion); ok {
		return v.ver, true
	}

	return "", false
}
