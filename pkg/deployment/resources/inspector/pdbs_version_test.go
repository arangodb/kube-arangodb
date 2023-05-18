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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	kversion "k8s.io/apimachinery/pkg/version"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Test_PDB_Versions(t *testing.T) {
	k8sVersions := map[string]version.Version{
		"v1.18.0": "",
		"v1.19.0": "",
		"v1.20.0": "",
		"v1.21.0": version.V1,
		"v1.22.0": version.V1,
		"v1.23.0": version.V1,
		"v1.24.0": version.V1,
		"v1.25.0": version.V1,
		"v1.26.0": version.V1,
	}

	for v, expected := range k8sVersions {
		t.Run(v, func(t *testing.T) {
			c := kclient.NewFakeClientWithVersion(&kversion.Info{
				GitVersion: v,
			})

			tc := throttle.NewThrottleComponents(time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour)

			i := NewInspector(tc, c, "test", "test")
			require.NoError(t, i.Refresh(context.Background()))

			if expected.IsV1() {
				_, err := i.PodDisruptionBudget().V1()
				require.NoError(t, err)
			} else {
				_, err := i.PodDisruptionBudget().V1()
				require.EqualError(t, err, "Kubernetes 1.20 or lower is not supported anymore")
			}
		})
	}
}
