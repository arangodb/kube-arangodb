//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v1alpha

// LocalStorageState is a strongly typed state of a deploymenlocal storage
type LocalStorageState string

const (
	// LocalStorageStateNone indicates that the state is not set yet
	LocalStorageStateNone LocalStorageState = ""
	// LocalStorageStateCreating indicates that the local storage components are being created
	LocalStorageStateCreating LocalStorageState = "Creating"
	// LocalStorageStateRunning indicates that all components are running
	LocalStorageStateRunning LocalStorageState = "Running"
	// LocalStorageStateFailed indicates that a local storage is in a failed state
	// from which automatic recovery is impossible. Inspect `Reason` for more info.
	LocalStorageStateFailed LocalStorageState = "Failed"
)

// IsFailed returns true if given state is LocalStorageStateFailed
func (cs LocalStorageState) IsFailed() bool {
	return cs == LocalStorageStateFailed
}
