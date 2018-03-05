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

package provisioner

import "context"

// API of the provisioner
type API interface {
	// GetFSInfo fetches information from the filesystem containing
	// the given local path.
	GetFSInfo(ctx context.Context, localPath string) (FSInfo, error)
	// Prepare a volume at the given local path
	Prepare(ctx context.Context, localPath string) error
	// Remove a volume with the given local path
	Remove(ctx context.Context, localPath string) error
}

// FSInfo holds information of a filesystem.
type FSInfo struct {
	Available int64 `json:"available"`
	Capacity  int64 `json:"capacity"`
}

// Request body for API HTTP requests.
type Request struct {
	LocalPath string `json:"localPath"`
}
