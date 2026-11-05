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

package arango

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Test_Platform_RegistersWorkflow(t *testing.T) {
	f := shared.NewFactoryGen()
	Platform(f)

	var names []string
	for _, factory := range f.Get() {
		names = append(names, factory.Name())
	}

	require.Contains(t, names, "platform-workflow")
	// The pre-existing platform collectors must remain registered.
	require.Contains(t, names, "platform-service")
	require.Contains(t, names, "platform-chart")
	require.Contains(t, names, "platform-storage")
}

func Test_Platform_WorkflowList(t *testing.T) {
	ctx := context.Background()
	client := kclient.NewFakeClient()
	ns := "test"

	_, err := client.Arango().PlatformV1beta1().ArangoPlatformWorkflows(ns).Create(ctx, &platformApi.ArangoPlatformWorkflow{
		ObjectMeta: meta.ObjectMeta{
			Name:      "wf1",
			Namespace: ns,
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	items, err := arangoPlatformV1beta1ArangoPlatformWorkflowList(ctx, client, ns)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "wf1", items[0].GetName())
}
