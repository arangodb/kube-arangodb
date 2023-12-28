//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package fwatch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func Test_FileContentWatcher(t *testing.T) {
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, uuid.New().String())
	data := []byte("12345\n6789")
	err := os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)
	defer os.Remove(filePath)

	fileWatcher, err := NewFileContentWatcher(filePath, logging.Global().Get("test"))
	require.NoError(t, err)
	w := fileWatcher.(*fileContentWatcher)

	ctx, cancel := context.WithCancel(context.Background())

	w.Start(ctx)
	require.True(t, w.isRunning)

	// must be initialized as "changed":
	time.Sleep(time.Millisecond * 100)
	require.True(t, w.IsChanged())

	// must read correctly and "changed" updated
	dataAct, err := w.ReadAll()
	require.NoError(t, err)
	require.Equal(t, data, dataAct)
	require.False(t, w.IsChanged())

	// change content
	newData := []byte("9876543210")
	err = os.WriteFile(filePath, newData, 0644)
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	require.True(t, w.IsChanged())

	// read and compare again
	dataAct, err = w.ReadAll()
	require.NoError(t, err)
	require.Equal(t, newData, dataAct)
	require.False(t, w.IsChanged())

	// recreate file with old data
	err = os.Remove(filePath)
	require.NoError(t, err)
	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	require.True(t, w.IsChanged())

	dataAct, err = w.ReadAll()
	require.NoError(t, err)
	require.Equal(t, data, dataAct)
	require.False(t, w.IsChanged())

	// cancel context: watchers should be stopped
	cancel()
	time.Sleep(time.Millisecond * 100)
	require.False(t, w.isRunning)
}
