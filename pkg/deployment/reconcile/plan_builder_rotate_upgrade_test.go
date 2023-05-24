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

package reconcile

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_RotateUpgrade_Condition(t *testing.T) {
	type testCase struct {
		status api.MemberStatus
		spec   api.DeploymentSpec
		images api.ImageInfoList

		verify func(t *testing.T, decision upgradeDecision)
	}

	newImageInfo := func(image string, imageID string, version driver.Version, enterprise bool) api.ImageInfo {
		return api.ImageInfo{
			Image:           image,
			ImageID:         imageID,
			ArangoDBVersion: version,
			Enterprise:      enterprise,
		}
	}
	newImageInfoP := func(image string, imageID string, version driver.Version, enterprise bool) *api.ImageInfo {
		p := newImageInfo(image, imageID, version, enterprise)
		return &p
	}

	newSpec := func(image string, mode api.DeploymentImageDiscoveryModeSpec) api.DeploymentSpec {
		return api.DeploymentSpec{
			Image:              util.NewType[string](image),
			ImageDiscoveryMode: api.NewDeploymentImageDiscoveryModeSpec(mode),
		}
	}

	testCases := map[string]testCase{
		"Unknown spec image - wait for discovery": {
			spec:   newSpec("unknown", api.DeploymentImageDiscoveryKubeletMode),
			images: api.ImageInfoList{}.Add(newImageInfo("a", "aid", "3.7.0", true)),

			verify: func(t *testing.T, decision upgradeDecision) {
				require.True(t, decision.Hold)
			},
		},
		"Missing image info": {
			spec:   newSpec("a", api.DeploymentImageDiscoveryKubeletMode),
			images: api.ImageInfoList{}.Add(newImageInfo("a", "aid", "3.7.0", true)),

			verify: func(t *testing.T, decision upgradeDecision) {
				require.False(t, decision.UpgradeNeeded)
			},
		},
		"Upgrade Kubelet case": {
			spec: newSpec("b", api.DeploymentImageDiscoveryKubeletMode),
			status: api.MemberStatus{
				Image: newImageInfoP("a", "aid", "3.7.0", true),
			},
			images: api.ImageInfoList{}.Add(newImageInfo("a", "aid", "3.7.0", true), newImageInfo("b", "bid", "3.8.0", true)),

			verify: func(t *testing.T, decision upgradeDecision) {
				require.True(t, decision.UpgradeNeeded)
			},
		},
		"Upgrade Direct case": {
			spec: newSpec("b", api.DeploymentImageDiscoveryDirectMode),
			status: api.MemberStatus{
				Image: newImageInfoP("a", "aid", "3.7.0", true),
			},
			images: api.ImageInfoList{}.Add(newImageInfo("a", "aid", "3.7.0", true), newImageInfo("b", "bid", "3.8.0", true)),

			verify: func(t *testing.T, decision upgradeDecision) {
				require.True(t, decision.UpgradeNeeded)
			},
		},
	}

	r := newTestReconciler()

	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			c.verify(t, r.podNeedsUpgrading(c.status, c.spec, c.images))
		})
	}
}
