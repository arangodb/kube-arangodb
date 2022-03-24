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

package upgrade

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func init() {
	registerUpgrade(memberCIDAppend())
}

func memberCIDAppend() Upgrade {
	return newUpgrade(api.Version{
		Major: 1,
		Minor: 2,
		Patch: 8,
		ID:    1,
	}, func(obj api.ArangoDeployment, status *api.DeploymentStatus, _ interfaces.Inspector) error {
		for _, i := range status.Members.AsList() {
			if i.Member.ClusterID == "" {
				i.Member.ClusterID = obj.GetUID()
				if err := status.Members.Update(i.Member, i.Group); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
