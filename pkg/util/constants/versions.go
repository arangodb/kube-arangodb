//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

const (
	VersionV1Alpha1 = "v1alpha1"
	VersionV1Beta1  = "v1beta1"
	VersionV1       = "v1"

	VersionV2Alpha1 = "v2alpha1"
)

func IsCompatible(base, version string) bool {
	return (IsV1Compatible(base) && IsV1Compatible(version)) || (IsV2Compatible(base) && IsV2Compatible(version))
}

func IsV1Compatible(version string) bool {
	return version == VersionV1Alpha1 ||
		version == VersionV1Beta1 ||
		version == VersionV1
}

func IsV2Compatible(version string) bool {
	return version == VersionV2Alpha1
}
