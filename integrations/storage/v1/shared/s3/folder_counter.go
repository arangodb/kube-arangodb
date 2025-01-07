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

package s3

import "strings"

// BucketFolderCounter is able to count "folder"s within buckets.
type BucketFolderCounter struct {
	basePath string
	folders  map[string]interface{}
}

// NewBucketFolderCounter creates a new BucketFolderCounter.
func NewBucketFolderCounter(basePath string) BucketFolderCounter {
	return BucketFolderCounter{
		basePath: basePath,
		folders:  make(map[string]interface{}),
	}
}

// AddObject extracts the folder names in the passed object name and adds them into the folders map.
func (b BucketFolderCounter) AddObject(obj string) {
	parts := strings.Split(obj, "/")
	for i := 0; i < len(parts)-1; i++ { // `parts-1` to not include actual file name
		// "folder1/folder2/file1" and "folder3/folder2/file2" are 4 different folders:
		// ["folder1", "folder1/folder2", "folder3", "folder3/folder2"]
		b.folders[strings.Join(parts[:i+1], "/")] = nil
	}
}

// GetFolderCount returns the number of folders within a bucket. Excludes the basePath.
func (b BucketFolderCounter) GetFolderCount() int32 {
	result := len(b.folders)

	if len(b.basePath) > 0 {
		rootFolders := strings.Split(b.basePath, "/")
		result -= len(rootFolders)
	}

	// This is possible when the basePath is file itself:
	// basePath = "folder1/file.txt"
	// AddObject("folder1/file.txt")
	if result < 0 {
		return 0
	}
	return int32(result)
}
