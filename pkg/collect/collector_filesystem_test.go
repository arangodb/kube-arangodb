//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

//go:build linux

package collect

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestFilesystemCollector(t *testing.T) {
	// The temp dir is backed by a real filesystem, so statfs must return a positive total and a
	// free value that never exceeds the total.
	values := collectValues(t, filesystemCollector{path: t.TempDir()})

	require.Greater(t, values[metricFSTotal], float32(0))
	require.GreaterOrEqual(t, values[metricFSFree], float32(0))
	require.LessOrEqual(t, values[metricFSFree], values[metricFSTotal])
}

func TestFilesystemCollector_MissingPath(t *testing.T) {
	// statfs on a non-existent path fails, and that failure propagates so the collection is retried.
	out := util.NewCollector[Metric]()
	err := filesystemCollector{path: "/this/path/does/not/exist/anywhere"}.CollectEvents(out)
	require.Error(t, err)
}
