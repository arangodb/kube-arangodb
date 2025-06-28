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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient/external"
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

func newClient(t *testing.T) (kclient.Client, string, Client) {
	client, ns := external.ExternalClient(t)

	c, err := NewClient(Configuration{
		Namespace: ns,
		Config:    client.Config(),
	})
	require.NoError(t, err)
	return client, ns, c
}
