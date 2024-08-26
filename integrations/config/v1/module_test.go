//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
)

func Test_Modules_Single(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	modules, err := client.Modules(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)
	require.Len(t, modules.GetModules(), 1)
}

func Test_Modules_Multi(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
			"test2": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	modules, err := client.Modules(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)
	require.Len(t, modules.GetModules(), 2)
}
