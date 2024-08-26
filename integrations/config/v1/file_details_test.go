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

func Test_Files_Details_Missing(t *testing.T) {
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

	_, err := client.FileDetails(ctx, &pbConfigV1.ConfigV1FileDetailsRequest{
		Module: "test",
		File:   "non-existent",
	})
	tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound).Errorf(t, "File `non-existent` not found within module `test`")
}

func Test_Files_Details_Empty(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).FileR(t, "file", 0)

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	resp, err := client.FileDetails(ctx, &pbConfigV1.ConfigV1FileDetailsRequest{
		Module: "test",
		File:   "file",
	})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetModule())
	require.NotNil(t, resp.GetFile())
	require.Equal(t, "file", resp.GetFile().GetPath())
	require.EqualValues(t, 0, resp.GetFile().GetSize())
	require.Empty(t, resp.GetFile().GetChecksum())
}

func Test_Files_Details_Empty_WC(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).FileR(t, "file", 0)

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	resp, err := client.FileDetails(ctx, &pbConfigV1.ConfigV1FileDetailsRequest{
		Module:   "test",
		File:     "file",
		Checksum: util.NewType(true),
	})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetModule())
	require.NotNil(t, resp.GetFile())
	require.Equal(t, "file", resp.GetFile().GetPath())
	require.EqualValues(t, 0, resp.GetFile().GetSize())
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", resp.GetFile().GetChecksum())
}

func Test_Files_Details_Data(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).File(t, "file", []byte("DATA"))

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	resp, err := client.FileDetails(ctx, &pbConfigV1.ConfigV1FileDetailsRequest{
		Module: "test",
		File:   "file",
	})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetModule())
	require.NotNil(t, resp.GetFile())
	require.Equal(t, "file", resp.GetFile().GetPath())
	require.EqualValues(t, 4, resp.GetFile().GetSize())
	require.Empty(t, resp.GetFile().GetChecksum())
}

func Test_Files_Details_Data_WC(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	dir := t.TempDir()

	tests.NewFileGenerator(t, dir).File(t, "file", []byte("DATA"))

	client := Client(t, ctx, Config{
		Modules: ModuleDefinitions{
			"test": {
				Path: dir,
			},
		},
	})

	require.NotNil(t, client)

	resp, err := client.FileDetails(ctx, &pbConfigV1.ConfigV1FileDetailsRequest{
		Module:   "test",
		File:     "file",
		Checksum: util.NewType(true),
	})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetModule())
	require.NotNil(t, resp.GetFile())
	require.Equal(t, "file", resp.GetFile().GetPath())
	require.EqualValues(t, 4, resp.GetFile().GetSize())
	require.Equal(t, "c97c29c7a71b392b437ee03fd17f09bb10b75e879466fc0eb757b2c4a78ac938", resp.GetFile().GetChecksum())
}
