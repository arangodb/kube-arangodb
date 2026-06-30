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

package collect

import (
	"os"
	"runtime"
	"strconv"
	goStrings "strings"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	// metricCPU is the number of CPUs available to the process.
	metricCPU = "system_cpu"

	// metricMemory is the total system memory in GB.
	metricMemory = "system_memory"

	// bytesPerGB is the number of bytes in one gigabyte.
	bytesPerGB = 1024 * 1024 * 1024
)

func init() {
	GetCollector().Register(resourceCollector{})
}

// resourceCollector pushes the available CPU and memory as event body metrics.
type resourceCollector struct{}

// CollectEvents pushes the CPU count and total memory (in GB).
func (resourceCollector) CollectEvents(out util.Pusher[Metric]) error {
	out.Push(Metric{K: metricCPU, V: float32(runtime.NumCPU())})

	mem, err := totalMemory()
	if err != nil {
		return err
	}

	out.Push(Metric{K: metricMemory, V: float32(mem) / bytesPerGB})

	return nil
}

// totalMemory returns the total system memory in bytes, read from /proc/meminfo.
func totalMemory() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, errors.Wrapf(err, "unable to read /proc/meminfo")
	}

	for _, line := range goStrings.Split(string(data), "\n") {
		if !goStrings.HasPrefix(line, "MemTotal:") {
			continue
		}

		// MemTotal is reported in kB, e.g. "MemTotal:       16384000 kB".
		fields := goStrings.Fields(line)
		if len(fields) < 2 {
			return 0, errors.Errorf("unexpected MemTotal line: %q", line)
		}

		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "unable to parse MemTotal: %q", line)
		}

		return kb * 1024, nil
	}

	return 0, errors.Errorf("MemTotal not found in /proc/meminfo")
}
