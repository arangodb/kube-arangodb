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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// collectValues runs a collector and returns its pushed metrics as a map.
func collectValues(t *testing.T, c ECollector[Metric]) map[string]float32 {
	t.Helper()
	out := util.NewCollector[Metric]()
	require.NoError(t, c.CollectEvents(out))
	require.NoError(t, out.Done())

	values := map[string]float32{}
	for _, m := range out.Collect() {
		values[m.K] = m.V
	}
	return values
}

func TestCgroupCollector_V2(t *testing.T) {
	root := t.TempDir()
	// Presence of cgroup.controllers marks the unified (v2) hierarchy.
	writeFile(t, filepath.Join(root, "cgroup.controllers"), "cpu memory")
	writeFile(t, filepath.Join(root, "cpu.max"), "50000 100000\n")  // 0.5 CPUs
	writeFile(t, filepath.Join(root, "memory.max"), "2147483648\n") // 2 GiB

	values := collectValues(t, cgroupCollector{root: root})

	require.InDelta(t, 0.5, values[metricCPULimit], 1e-6)
	require.InDelta(t, 2.0, values[metricMemoryLimit], 1e-6)
}

func TestCgroupCollector_V2_Unlimited(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "cgroup.controllers"), "cpu memory")
	writeFile(t, filepath.Join(root, "cpu.max"), "max 100000\n")
	writeFile(t, filepath.Join(root, "memory.max"), "max\n")

	values := collectValues(t, cgroupCollector{root: root})

	// Unlimited means nothing is pushed.
	_, hasCPU := values[metricCPULimit]
	_, hasMem := values[metricMemoryLimit]
	require.False(t, hasCPU)
	require.False(t, hasMem)
}

func TestCgroupCollector_V1(t *testing.T) {
	root := t.TempDir()
	// No cgroup.controllers file -> v1 hierarchy.
	writeFile(t, filepath.Join(root, "cpu", "cpu.cfs_quota_us"), "150000\n") // 1.5 CPUs
	writeFile(t, filepath.Join(root, "cpu", "cpu.cfs_period_us"), "100000\n")
	writeFile(t, filepath.Join(root, "memory", "memory.limit_in_bytes"), "1073741824\n") // 1 GiB

	values := collectValues(t, cgroupCollector{root: root})

	require.InDelta(t, 1.5, values[metricCPULimit], 1e-6)
	require.InDelta(t, 1.0, values[metricMemoryLimit], 1e-6)
}

func TestCgroupCollector_V1_Unlimited(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "cpu", "cpu.cfs_quota_us"), "-1\n")
	writeFile(t, filepath.Join(root, "cpu", "cpu.cfs_period_us"), "100000\n")
	// The v1 "no limit" sentinel.
	writeFile(t, filepath.Join(root, "memory", "memory.limit_in_bytes"), "9223372036854771712\n")

	values := collectValues(t, cgroupCollector{root: root})

	_, hasCPU := values[metricCPULimit]
	_, hasMem := values[metricMemoryLimit]
	require.False(t, hasCPU)
	require.False(t, hasMem)
}

func TestCgroupCollector_Absent(t *testing.T) {
	// An empty root has none of the cgroup files: the collector must succeed and push nothing,
	// rather than fail the whole collection (e.g. when running outside a container).
	values := collectValues(t, cgroupCollector{root: t.TempDir()})
	require.Empty(t, values)
}
