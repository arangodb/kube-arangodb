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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketFolderCounter(t *testing.T) {
	t.Run("base path same with the object, should return 0", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path/file1.txt")
		b.AddObject("my_root_path/file1.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(0), numFolders)
	})

	t.Run("no folder within base path, should return 0", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path")
		b.AddObject("my_root_path/file1.txt")
		b.AddObject("my_root_path/file2.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(0), numFolders)
	})

	t.Run("[folder1, folder2, folder1/folder2, folder1/folder2/folder2] should return 4", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path")
		b.AddObject("my_root_path/file1.txt")
		b.AddObject("my_root_path/file2.txt")
		b.AddObject("my_root_path/folder1/file1.txt")
		b.AddObject("my_root_path/folder2/file1.txt")
		b.AddObject("my_root_path/folder1/folder2/file1.txt")
		b.AddObject("my_root_path/folder1/folder2/folder2/file1.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(4), numFolders)
	})

	t.Run("[folder1, folder3, folder1/folder2, folder3/folder2] should return 4", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path")
		b.AddObject("my_root_path/folder1/folder2/file1.txt")
		b.AddObject("my_root_path/folder3/folder2/file2.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(4), numFolders)
	})

	t.Run("nested is part of the base path, should return 0", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path/nested")
		b.AddObject("my_root_path/nested/file1.txt")
		b.AddObject("my_root_path/nested/file2.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(0), numFolders)
	})

	t.Run("nested should be excluded. [folder1, folder2, folder1/folder2, folder1/folder2/folder2] should return 4", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path/nested")
		b.AddObject("my_root_path/nested/file1.txt")
		b.AddObject("my_root_path/nested/file2.txt")
		b.AddObject("my_root_path/nested/folder1/file1.txt")
		b.AddObject("my_root_path/nested/folder2/file1.txt")
		b.AddObject("my_root_path/nested/folder1/folder2/file1.txt")
		b.AddObject("my_root_path/nested/folder1/folder2/folder2/file1.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(4), numFolders)
	})

	t.Run("nested should be excluded. [folder1, folder3, folder1/folder2, folder3/folder2] should return 4", func(t *testing.T) {
		b := NewBucketFolderCounter("my_root_path/nested")
		b.AddObject("my_root_path/nested/folder1/folder2/file1.txt")
		b.AddObject("my_root_path/nested/folder3/folder2/file2.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(4), numFolders)
	})

	t.Run("counting for bucket root should work", func(t *testing.T) {
		b := NewBucketFolderCounter("")
		b.AddObject("my_root_path/file1.txt")
		numFolders := b.GetFolderCount()
		assert.Equal(t, int32(1), numFolders)
	})
}
