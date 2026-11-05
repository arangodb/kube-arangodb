//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package platform

import (
	"testing"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/cache"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const testNamespace = "default"

func mlChart(namespace, version string) *platformApi.ArangoPlatformChart {
	return &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{Name: "ml", Namespace: namespace},
		Status: platformApi.ArangoPlatformChartStatus{
			Info: &platformApi.ChartStatusInfo{
				Details: &platformApi.ChartDetails{Name: "ml", Version: version},
			},
		},
	}
}

func deploymentWithArangoDBVersion(version string) *api.ArangoDeployment {
	return &api.ArangoDeployment{
		Status: api.DeploymentStatus{
			CurrentImage: &api.ImageInfo{ArangoDBVersion: adbDriverV2.Version(version)},
		},
	}
}

func newCharts(charts ...*platformApi.ArangoPlatformChart) cache.ObjectCache[*platformApi.ArangoPlatformChart] {
	c := NewCache(0)
	for _, chart := range charts {
		c.Charts().Set(operation.ItemName{Namespace: chart.GetNamespace(), Name: chart.GetName()}, chart)
	}
	return c.Charts()
}

func Test_EvaluateRequirements(t *testing.T) {
	t.Run("No requirements", func(t *testing.T) {
		failures := EvaluateRequirements(newCharts(), testNamespace, nil, nil)
		require.Empty(t, failures)
	})

	t.Run("Empty constraint is skipped", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain("")}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, nil)
		require.Empty(t, failures)
	})

	t.Run("Custom chart satisfied", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain(">=1.2.0")}
		failures := EvaluateRequirements(newCharts(mlChart(testNamespace, "1.3.0")), testNamespace, reqs, nil)
		require.Empty(t, failures)
	})

	t.Run("Custom chart version does not satisfy", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain(">=1.2.0")}
		failures := EvaluateRequirements(newCharts(mlChart(testNamespace, "1.0.0")), testNamespace, reqs, nil)
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], "ml:")
		require.Contains(t, failures[0], "does not satisfy")
	})

	t.Run("Custom chart not available", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain(">=1.2.0")}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, nil)
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], "ml: chart is not available")
	})

	t.Run("Custom chart version not known yet", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain(">=1.2.0")}
		failures := EvaluateRequirements(newCharts(mlChart(testNamespace, "")), testNamespace, reqs, nil)
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], "ml: version is not known yet")
	})

	t.Run("Custom chart in another namespace is not visible", func(t *testing.T) {
		other := mlChart("other", "1.3.0")
		reqs := platformApi.ChartDetailsPlatformRequirements{"ml": util.VersionConstrain(">=1.2.0")}
		failures := EvaluateRequirements(newCharts(other), testNamespace, reqs, nil)
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], "ml: chart is not available")
	})

	t.Run("Reserved ArangoDB satisfied", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{platformApi.ChartRequirementArangoDB: util.VersionConstrain(">=3.12.0")}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, deploymentWithArangoDBVersion("3.12.1"))
		require.Empty(t, failures)
	})

	t.Run("Reserved ArangoDB does not satisfy", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{platformApi.ChartRequirementArangoDB: util.VersionConstrain(">=3.12.0")}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, deploymentWithArangoDBVersion("3.11.0"))
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], platformApi.ChartRequirementArangoDB+":")
	})

	t.Run("Reserved ArangoDB version not known", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{platformApi.ChartRequirementArangoDB: util.VersionConstrain(">=3.12.0")}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, nil)
		require.Len(t, failures, 1)
		require.Contains(t, failures[0], "version is not known yet")
	})

	t.Run("Multiple custom failures are sorted", func(t *testing.T) {
		reqs := platformApi.ChartDetailsPlatformRequirements{
			"zeta":  util.VersionConstrain(">=2.0.0"),
			"alpha": util.VersionConstrain(">=2.0.0"),
		}
		failures := EvaluateRequirements(newCharts(), testNamespace, reqs, nil)
		require.Len(t, failures, 2)
		require.Contains(t, failures[0], "alpha:")
		require.Contains(t, failures[1], "zeta:")
	})
}
