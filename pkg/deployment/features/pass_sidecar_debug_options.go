//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package features

func init() {
	registerFeature(passSidecarDebugOptions)
}

var passSidecarDebugOptions = &feature{
	name:             "pass-sidecar-debug-options",
	description:      "Propagate the operator's own --log.level options to the integration and internal authorization sidecars (diagnostics only)",
	enabledByDefault: false,
	hidden:           true,
}

// PassSidecarDebugOptions reports whether the operator's own --log.level options should be propagated to the
// integration and internal authorization sidecars, so a single operator-level setting (e.g. --log.level=debug,
// or --log.level=grpc-service=debug for just the sidecar->central proxy trace) turns on sidecar debug logging
// too. Off by default; intended for diagnostics.
func PassSidecarDebugOptions() Feature {
	return passSidecarDebugOptions
}
