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

package compare

import (
	"encoding/json"

	jd "github.com/josephburnett/jd/lib"
)

func Diff[T interface{}](spec, status *T) (specData, statusData, diff string, outErr error) {
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return "", "", "", err
	}
	specData = string(specBytes)

	statusBytes, err := json.Marshal(status)
	if err != nil {
		return "", "", "", err
	}
	statusData = string(statusBytes)

	if specData, err := jd.ReadJsonString(string(specBytes)); err != nil {
		return "", "", "", err
	} else if specData != nil {
		if statusData, err := jd.ReadJsonString(string(statusBytes)); err != nil {
			return "", "", "", err
		} else if statusData != nil {
			diff = specData.Diff(statusData).Render()
		}
	}

	return
}
