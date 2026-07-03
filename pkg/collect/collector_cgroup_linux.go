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
	"strconv"
	goStrings "strings"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// The resource collector in collector_resource.go reports host-level values (runtime.NumCPU and
// /proc/meminfo). Inside a container those over-report: they reflect the node, not the cgroup limit
// the customer was actually allocated. This collector reports the cgroup-enforced limits instead,
// supporting both cgroup v1 (per-controller hierarchy) and cgroup v2 (unified hierarchy).
//
// The cgroup filesystem is a Linux-only kernel interface, so this collector is compiled and
// registered on Linux only. See collector_cgroup_other.go for the no-op on other platforms.
//
// When no limit is enforced (unlimited) or the cgroup files are not present (e.g. running outside a
// container), the corresponding metric is simply not pushed - this is not an error, the collector
// just has nothing to report.

const (
	// metricCPULimit is the number of CPUs the cgroup allows (fractional, e.g. quota/period).
	metricCPULimit = "system_cpu_limit"

	// metricMemoryLimit is the cgroup memory limit in GB.
	metricMemoryLimit = "system_memory_limit"

	// defaultCgroupRoot is the standard mount point of the cgroup filesystem.
	defaultCgroupRoot = "/sys/fs/cgroup"

	// unlimitedMemoryThreshold guards against the cgroup v1 "no limit" sentinel, which is reported
	// as a page-aligned near-max int64 (e.g. 9223372036854771712). Any value at or above this is
	// treated as unlimited rather than a real limit.
	unlimitedMemoryThreshold = uint64(1) << 62
)

func init() {
	GetCollector().Register(cgroupCollector{root: defaultCgroupRoot})
}

// cgroupCollector pushes the cgroup-enforced CPU and memory limits as event body metrics.
type cgroupCollector struct {
	// root is the cgroup filesystem mount point. Overridable for testing.
	root string
}

// CollectEvents pushes the cgroup CPU limit (in CPUs) and memory limit (in GB), when enforced.
func (c cgroupCollector) CollectEvents(out util.Pusher[Metric]) error {
	root := c.root
	if root == "" {
		root = defaultCgroupRoot
	}

	v2 := isCgroupV2(root)

	cpu, ok, err := cgroupCPULimit(root, v2)
	if err != nil {
		return err
	}
	if ok {
		out.Push(Metric{K: metricCPULimit, V: cpu})
	}

	mem, ok, err := cgroupMemoryLimit(root, v2)
	if err != nil {
		return err
	}
	if ok {
		out.Push(Metric{K: metricMemoryLimit, V: float32(mem) / bytesPerGB})
	}

	return nil
}

// isCgroupV2 reports whether the cgroup mount is the unified (v2) hierarchy. The unified hierarchy
// is identified by the presence of the cgroup.controllers file at the root.
func isCgroupV2(root string) bool {
	_, err := os.Stat(filepath.Join(root, "cgroup.controllers"))
	return err == nil
}

// cgroupCPULimit returns the number of CPUs the cgroup permits (quota/period). ok is false when no
// limit is enforced or the relevant files are absent (running outside a container).
func cgroupCPULimit(root string, v2 bool) (float32, bool, error) {
	if v2 {
		// cgroup v2: cpu.max holds "<quota> <period>", or "max <period>" when unlimited.
		line, ok, err := readCgroupFile(filepath.Join(root, "cpu.max"))
		if err != nil || !ok {
			return 0, false, err
		}

		fields := goStrings.Fields(line)
		if len(fields) < 2 {
			return 0, false, errors.Errorf("unexpected cpu.max content: %q", line)
		}
		if fields[0] == "max" {
			return 0, false, nil
		}

		quota, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0, false, errors.Wrapf(err, "unable to parse cpu.max quota: %q", line)
		}
		period, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0, false, errors.Wrapf(err, "unable to parse cpu.max period: %q", line)
		}
		if period <= 0 {
			return 0, false, nil
		}
		return float32(quota) / float32(period), true, nil
	}

	// cgroup v1: cpu.cfs_quota_us / cpu.cfs_period_us. A quota of -1 means unlimited.
	quotaStr, ok, err := readCgroupFile(filepath.Join(root, "cpu", "cpu.cfs_quota_us"))
	if err != nil || !ok {
		return 0, false, err
	}
	quota, err := strconv.ParseInt(goStrings.TrimSpace(quotaStr), 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "unable to parse cpu.cfs_quota_us: %q", quotaStr)
	}
	if quota <= 0 {
		return 0, false, nil
	}

	periodStr, ok, err := readCgroupFile(filepath.Join(root, "cpu", "cpu.cfs_period_us"))
	if err != nil || !ok {
		return 0, false, err
	}
	period, err := strconv.ParseInt(goStrings.TrimSpace(periodStr), 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "unable to parse cpu.cfs_period_us: %q", periodStr)
	}
	if period <= 0 {
		return 0, false, nil
	}

	return float32(quota) / float32(period), true, nil
}

// cgroupMemoryLimit returns the cgroup memory limit in bytes. ok is false when no limit is enforced
// or the relevant files are absent.
func cgroupMemoryLimit(root string, v2 bool) (uint64, bool, error) {
	var file string
	if v2 {
		file = filepath.Join(root, "memory.max")
	} else {
		file = filepath.Join(root, "memory", "memory.limit_in_bytes")
	}

	line, ok, err := readCgroupFile(file)
	if err != nil || !ok {
		return 0, false, err
	}

	line = goStrings.TrimSpace(line)

	// cgroup v2 reports "max" for no limit.
	if line == "max" {
		return 0, false, nil
	}

	limit, err := strconv.ParseUint(line, 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "unable to parse memory limit: %q", line)
	}

	// cgroup v1 reports a near-max sentinel instead of "max".
	if limit >= unlimitedMemoryThreshold {
		return 0, false, nil
	}

	return limit, true, nil
}

// readCgroupFile reads a cgroup file, returning its trimmed content. ok is false (without error)
// when the file does not exist, so callers can treat a missing cgroup file as "no limit reported".
func readCgroupFile(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, errors.Wrapf(err, "unable to read %q", path)
	}
	return goStrings.TrimSpace(string(data)), true, nil
}
