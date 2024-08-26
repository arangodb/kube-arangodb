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
	"google.golang.org/grpc/codes"

	pbConfigV1 "github.com/arangodb/kube-arangodb/integrations/config/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Modules_Details_Empty(t *testing.T) {
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

	_, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{})
	tgrpc.AsGRPCError(t, err).Code(t, codes.InvalidArgument).Errorf(t, "Module name cannot be empty")
}

func Test_Modules_Details_NotFound(t *testing.T) {
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

	_, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{
		Module: "some",
	})
	tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound).Errorf(t, "Module `some` not found")
}

func Test_Modules_Details_Exists_EmptyFiles(t *testing.T) {
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

	module, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{
		Module: "test",
	})
	require.NoError(t, err)
	require.Equal(t, "test", module.GetModule())
	require.Len(t, module.GetFiles(), 0)
	require.Equal(t, "", module.GetChecksum())
}

func Test_Modules_Details_Exists_EmptyFiles_WithChecksum(t *testing.T) {
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

	module, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{
		Module:   "test",
		Checksum: util.NewType(true),
	})
	require.NoError(t, err)
	require.Equal(t, "test", module.GetModule())
	require.Len(t, module.GetFiles(), 0)
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", module.GetChecksum())
}

func Test_Modules_Details_Exists_SomeFiles(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).
		FileR(t, "test", 128).
		Directory(t, "sub").FileR(t, "test", 128)

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	module, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{
		Module: "test",
	})
	require.NoError(t, err)
	require.Equal(t, "test", module.GetModule())
	require.Len(t, module.GetFiles(), 2)
	files := module.GetFiles()
	require.Equal(t, "sub/test", files[0].GetPath())
	require.Equal(t, "test", files[1].GetPath())
	require.Equal(t, "", module.GetChecksum())
}

func Test_Modules_Details_Exists_SomeFiles_WithChecksum(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).
		File(t, "test", []byte("DATA")).
		Directory(t, "sub").File(t, "test", []byte("DATA2"))

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	module, err := client.ModuleDetails(ctx, &pbConfigV1.ConfigV1ModuleDetailsRequest{
		Module:   "test",
		Checksum: util.NewType(true),
	})
	require.NoError(t, err)
	require.Equal(t, "test", module.GetModule())
	require.Len(t, module.GetFiles(), 2)
	files := module.GetFiles()
	require.Equal(t, "sub/test", files[0].GetPath())
	require.Equal(t, "test", files[1].GetPath())
	require.Equal(t, "e357414aec56cf8e5e3988b53b766049521a1f4920ad72462d48ebbe16942915", module.GetChecksum())
}
