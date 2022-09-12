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

package deployment

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

type RecoverStatusFunc func(in *api.DeploymentStatus) (bool, error)

func RecoverStatus(in *api.DeploymentStatus, fs ...RecoverStatusFunc) (bool, error) {
	var changed bool
	for _, f := range fs {
		if f != nil {
			if c, err := f(in); err != nil {
				return false, err
			} else if c {
				changed = true
			}
		}
	}
	return changed, nil
}
