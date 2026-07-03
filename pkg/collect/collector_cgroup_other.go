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

//go:build !linux

package collect

// The cgroup collector (collector_cgroup_linux.go) reads the Linux cgroup filesystem
// (/sys/fs/cgroup), a Linux-only kernel interface. On non-Linux platforms it is intentionally not
// compiled or registered, so the package still builds for local development and tooling. No cgroup
// resource metrics are emitted off Linux.
