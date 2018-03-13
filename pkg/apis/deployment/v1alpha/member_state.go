//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

// MemberState is a strongly typed state of a deployment member
type MemberState string

const (
	// MemberStateNone indicates that the state is not set yet
	MemberStateNone MemberState = ""
	// MemberStateCreated indicates that all resources needed for the member have been created
	MemberStateCreated MemberState = "Created"
	// MemberStateCleanOut indicates that a dbserver is in the process of being cleaned out
	MemberStateCleanOut MemberState = "CleanOut"
	// MemberStateShuttingDown indicates that a member is shutting down
	MemberStateShuttingDown MemberState = "ShuttingDown"
)
