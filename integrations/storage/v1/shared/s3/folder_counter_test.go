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
