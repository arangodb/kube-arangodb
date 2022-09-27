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

package rotation

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func init() {
	k8sutil.SetBinaryPath("arangodb_operator")
}

//go:embed testdata/pod_lifecycle_change.000.spec.json
var podLifecycleChange000Spec []byte

//go:embed testdata/pod_lifecycle_change.000.status.json
var podLifecycleChange000Status []byte

func runPredefinedTests(t *testing.T, spec, status []byte) (mode Mode, plan api.Plan, err error) {
	var specO, statusO core.PodTemplateSpec

	require.NoError(t, json.Unmarshal(spec, &specO))
	require.NoError(t, json.Unmarshal(status, &statusO))

	specC, err := resources.ChecksumArangoPod(api.ServerGroupSpec{}, resources.CreatePodFromTemplate(&specO))
	require.NoError(t, err)

	statusC, err := resources.ChecksumArangoPod(api.ServerGroupSpec{}, resources.CreatePodFromTemplate(&statusO))
	require.NoError(t, err)

	obj := api.DeploymentSpec{}
	member := api.MemberStatus{}

	specT, err := api.GetArangoMemberPodTemplate(&specO, specC)
	require.NoError(t, err)
	statusT, err := api.GetArangoMemberPodTemplate(&statusO, statusC)
	require.NoError(t, err)

	return compare(obj, member, api.ServerGroupUnknown, specT, statusT)
}

func Test_PredefinedTests(t *testing.T) {
	t.Run("podLifecycleChange000", func(t *testing.T) {
		mode, plan, err := runPredefinedTests(t, podLifecycleChange000Spec, podLifecycleChange000Status)
		require.NoError(t, err)
		require.Empty(t, plan)
		require.Equal(t, SilentRotation, mode)
	})
}
