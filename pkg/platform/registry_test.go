//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"
	"testing"

	"github.com/regclient/regclient"
	_ "github.com/regclient/regclient"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/platform/pack"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func Test_Registry_New(t *testing.T) {
	ctx := shutdown.Context()

	cm, err := helm.NewChartManager(ctx, goHttp.DefaultClient, "%s/index.yaml", "https://arangodb-platform-dev-chart-registry.s3.amazonaws.com")
	require.NoError(t, err)

	client := regclient.New(regclient.WithDockerCreds())

	require.NoError(t, pack.ExportPackage(ctx, "output2.zip", cm, client, helm.Package{
		Packages: map[string]helm.PackageSpec{
			"arangodb-platform-ui": {Version: "v0.1.1-1e61113"},
		},
	}))
}

func Test_Registry_Restore_New(t *testing.T) {
	ctx := shutdown.Context()

	client := regclient.New(regclient.WithDockerCreds())

	pkg, err := pack.ImportPackage(ctx, "output2.zip", client, "gcr.io/gcr-for-testing/aj/tr3")
	require.NoError(t, err)

	data, err := yaml.Marshal(pkg)
	require.NoError(t, err)

	println(string(data))
}
