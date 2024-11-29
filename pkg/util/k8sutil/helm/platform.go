//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package helm

import (
	"helm.sh/helm/v3/pkg/chart"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	PlatformFileName = "platform.yml"
)

func extractPlatform(chart *chart.Chart) (*Platform, error) {
	if chart == nil {
		return nil, nil
	}

	for _, file := range chart.Files {
		if file == nil {
			return nil, nil
		}

		if file.Name == PlatformFileName {
			obj, err := util.JsonOrYamlUnmarshal[Platform](file.Data)
			if err != nil {
				return nil, err
			}

			return &obj, nil
		}
	}

	return nil, nil
}
