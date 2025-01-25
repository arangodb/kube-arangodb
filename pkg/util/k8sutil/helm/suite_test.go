//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func init() {
	logging.Global().ApplyLogLevels(map[string]logging.Level{
		logging.TopicAll: logging.Debug,
	})
}

func newValues(t *testing.T, in any) Values {
	v, err := NewValues(in)
	require.NoError(t, err)
	return v
}

func newClient(t *testing.T, namespace string) (kclient.Client, Client) {
	z, ok := os.LookupEnv("TEST_KUBECONFIG")
	if !ok {
		t.Skipf("TEST_KUBECONFIG is not set")
	}

	kcfg, err := clientcmd.BuildConfigFromFlags("", z)
	require.NoError(t, err)

	client, err := kclient.NewClient("test", kcfg)
	require.NoError(t, err)

	c, err := NewClient(Configuration{
		Namespace: namespace,
		Config:    client.Config(),
	})
	require.NoError(t, err)
	return client, c
}
