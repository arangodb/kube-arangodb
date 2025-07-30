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

type Version string

const (
	VersionV1Alpha1 Version = "v1alpha1"
	VersionV1Beta1  Version = "v1beta1"
	VersionV1       Version = "v1"

	VersionV2Alpha1 Version = "v2alpha1"
)

func (v Version) String() string {
	return string(v)
}

func (v Version) Is(other Version) bool {
	return v == other
}

func (v Version) IsCompatible(other Version) bool {
	return (v.IsV1Compatible() && other.IsV1Compatible()) || (v.IsV2Compatible() && other.IsV2Compatible())
}

func (v Version) IsV1Compatible() bool {
	return v == VersionV1Alpha1 ||
		v == VersionV1Beta1 ||
		v == VersionV1
}

func (v Version) IsV2Compatible() bool {
	return v == VersionV2Alpha1
}
