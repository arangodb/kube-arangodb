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

package rotation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

func compareFunc(deploymentSpec api.DeploymentSpec, member api.MemberStatus, group api.ServerGroup,
	spec, status *api.ArangoMemberPodTemplate) (mode compare.Mode, plan api.Plan, err error) {
	return compare.P2[core.PodTemplateSpec, api.DeploymentSpec, api.ServerGroup](logger,
		deploymentSpec, group,
		actions.NewActionBuilderWrap(group, member),
		func(in *core.PodTemplateSpec) (string, error) {
			data, err := json.Marshal(in.Spec)
			if err != nil {
				return "", err
			}

			checksum := fmt.Sprintf("%0x", sha256.Sum256(data))

			return checksum, nil
		},
		spec, status,
		podCompare, affinityCompare, comparePodVolumes, containersCompare, initContainersCompare, comparePodTolerations)
}
