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

package features

func init() {
	registerFeature(localVolumeReplacementCheck)
	registerFeature(localStorageReclaimPolicyPass)
}

var localVolumeReplacementCheck Feature = &feature{
	name:               "local-volume-replacement-check",
	description:        "Replace volume for local-storage if volume is unschedulable (ex. node is gone)",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

var localStorageReclaimPolicyPass Feature = &feature{
	name:               "local-storage.pass-reclaim-policy",
	description:        "[LocalStorage] Pass ReclaimPolicy from StorageClass instead of using hardcoded Retain",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
}

func LocalStorageReclaimPolicyPass() Feature {
	return localStorageReclaimPolicyPass
}

func LocalVolumeReplacementCheck() Feature {
	return localVolumeReplacementCheck
}
