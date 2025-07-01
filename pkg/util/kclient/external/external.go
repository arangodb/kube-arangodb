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

//go:build testing

package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/arangodb/kube-arangodb/pkg/crd"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const TEST_KUBECONFIG util.EnvironmentVariable = "TEST_KUBECONFIG"

func ExternalClient(t *testing.T) (kclient.Client, string) {
	if !TEST_KUBECONFIG.Exists() {
		t.Skipf("TEST_KUBECONFIG is not set")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", TEST_KUBECONFIG.Get())
	require.NoError(t, err)

	cfg.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()

	client, err := kclient.NewClient("test", cfg)
	require.NoError(t, err)

	require.True(t, t.Run("Ensure CRDs", func(t *testing.T) {
		require.NoError(t, crd.EnsureCRDWithOptions(shutdown.Context(), client, crd.EnsureCRDOptions{}))
	}), "Unable to install CRDs")

	ns, err := client.Kubernetes().CoreV1().Namespaces().Create(shutdown.Context(), &core.Namespace{
		ObjectMeta: meta.ObjectMeta{
			GenerateName: "test-it-",
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := client.Kubernetes().CoreV1().Namespaces().Delete(context.Background(), ns.Name, meta.DeleteOptions{})
		require.NoError(t, err)
	})

	return client, ns.GetName()
}
