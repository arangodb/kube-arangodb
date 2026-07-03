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
	"golang.org/x/sys/unix"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// filesystemCollector reports the total and available capacity of the filesystem backing a given
// path, obtained from statfs. This is the disk-usage half of the resource inventory: how much
// storage the container has, and how much of it is still free.
//
// statfs is compiled on Linux only here; see collector_filesystem_other.go for the no-op on other
// platforms.

const (
	// metricFSTotal is the total capacity of the filesystem under the path, in GB.
	metricFSTotal = "system_fs_total"

	// metricFSFree is the space available to an unprivileged user on that filesystem, in GB.
	metricFSFree = "system_fs_free"

	// defaultFilesystemPath is the path whose backing filesystem is measured by default. The root
	// filesystem is a safe default; a real deployment would point this at the data volume.
	defaultFilesystemPath = "/"
)

func init() {
	GetCollector().Register(filesystemCollector{path: defaultFilesystemPath})
}

// filesystemCollector pushes the total and free capacity of the filesystem backing path.
type filesystemCollector struct {
	// path selects the filesystem to measure (any path on the target mount). Overridable for testing.
	path string
}

// CollectEvents pushes the total and available capacity (in GB) of the filesystem backing the path.
func (f filesystemCollector) CollectEvents(out util.Pusher[Metric]) error {
	path := f.path
	if path == "" {
		path = defaultFilesystemPath
	}

	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return errors.Wrapf(err, "unable to statfs %q", path)
	}

	// Bsize is the fundamental block size; Blocks is the total number of blocks and Bavail the
	// blocks available to an unprivileged user. Multiplying by the block size yields bytes.
	blockSize := uint64(st.Bsize)
	total := st.Blocks * blockSize
	free := st.Bavail * blockSize

	out.Push(Metric{K: metricFSTotal, V: float32(total) / bytesPerGB})
	out.Push(Metric{K: metricFSFree, V: float32(free) / bytesPerGB})

	return nil
}
