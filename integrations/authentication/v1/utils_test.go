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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type tokenType [32]byte

func generateJWTToken() tokenType {
	var tokenData tokenType
	util.Rand().Read(tokenData[:])
	return tokenData
}

func reSaveJWTTokens(t *testing.T, directory string, datas ...tokenType) {
	cleanJWTTokens(t, directory)
	saveJWTTokens(t, directory, datas...)
}

func cleanJWTTokens(t *testing.T, directory string) {
	files, err := os.ReadDir(directory)
	require.NoError(t, err)

	for _, f := range files {
		require.NoError(t, os.Remove(path.Join(directory, f.Name())))
	}

	files, err = os.ReadDir(directory)
	require.NoError(t, err)
	require.Len(t, files, 0)
}

func saveJWTTokens(t *testing.T, directory string, datas ...tokenType) {
	require.True(t, len(datas) > 0, "Required at least one token")
	saveJWTToken(t, directory, "-", datas[0])

	for _, data := range datas {
		saveJWTToken(t, directory, util.SHA256(data[:]), data)
	}
}

func saveJWTToken(t *testing.T, directory, name string, data tokenType) {
	fn := path.Join(directory, name)
	require.NoError(t, os.WriteFile(fn, data[:], 0644))
}
