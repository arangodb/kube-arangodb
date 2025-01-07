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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbStorage "github.com/arangodb-managed/integration-apis/bucket-service/v1"
	icommon "github.com/arangodb-managed/integration-apis/common/v1"
)

func Test_BucketServiceServerTestManualRun(t *testing.T) {
	testLargeFile := !testing.Short()

	s, err := NewS3Impl(getClient(t))

	tid := fmt.Sprintf("temp/%s", uuid.NewUUID())

	assert.NoError(t, err)

	server := s.(*s3impl)

	ctx := context.Background()
	assert.NotNil(t, ctx)

	rounding := time.Millisecond * 100
	assert.NotZero(t, rounding)
	maxTimeDiffServers := time.Second * 2
	assert.NotZero(t, maxTimeDiffServers)

	// Validate it's there
	existsResult, err := server.BucketExists(ctx, &pbStorage.BucketRequest{})
	assert.NoError(t, err)
	assert.True(t, existsResult.GetResult())
	// Get Size (should be 0)
	pSize, err := server.PathInfo(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFolders())
	// Add data
	fileTags := icommon.KeyValuePairList{
		{
			Key:   "ftestk",
			Value: "ftestv",
		},
	}
	// Capture current time (however we are not completely synced with the provider, so allow 2 seconds diff)
	dtBefore := time.Now().UTC().Add(-maxTimeDiffServers)
	moreDataAllowed, err := server.write(ctx, fmt.Sprintf("%s/temp/test.txt", tid), []byte("test"), true, fileTags)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	info, err := server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test.txt", tid))
	assert.NoError(t, err)
	dtAfter := time.Now().UTC().Add(maxTimeDiffServers)
	assert.True(t, info.Exists)
	assert.False(t, info.IsLocked)
	assert.Equal(t, uint64(4), info.SizeInBytes)
	assert.Len(t, info.Tags, 1)
	assert.Equal(t, "ftestv", *info.Tags.GetValue("ftestk"))
	firstLastUpdatedAt := info.LastUpdatedAt
	assert.WithinDuration(t, dtBefore, info.LastUpdatedAt, dtAfter.Sub(dtBefore), fmt.Sprintf("%v not between %v and %v", info.LastUpdatedAt, dtBefore, dtAfter))
	// Wait a little before appending more data
	time.Sleep(maxTimeDiffServers)
	// Append data
	dtBefore = time.Now().UTC().Add(-maxTimeDiffServers)
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test.txt", tid), []byte("123"), false, nil)
	assert.NoError(t, err)
	assert.False(t, moreDataAllowed)
	info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test.txt", tid))
	assert.NoError(t, err)
	dtAfter = time.Now().Add(maxTimeDiffServers)
	assert.True(t, info.Exists)
	assert.True(t, info.IsLocked)
	assert.Equal(t, uint64(7), info.SizeInBytes)
	assert.Len(t, info.Tags, 1)
	assert.Equal(t, "ftestv", *info.Tags.GetValue("ftestk"))
	assert.WithinDuration(t, dtBefore, info.LastUpdatedAt, dtAfter.Sub(dtBefore), fmt.Sprintf("%v not between %v and %v", info.LastUpdatedAt, dtBefore, dtAfter))
	assert.True(t, firstLastUpdatedAt.Before(info.LastUpdatedAt))
	// Try to append more data
	_, err = server.write(ctx, fmt.Sprintf("%s/temp/test.txt", tid), []byte("should-fail"), false, nil)
	assert.Error(t, err)
	// Wait a little before reading it again
	time.Sleep(maxTimeDiffServers)
	// Get Data
	reader, err := server.createReader(ctx, fmt.Sprintf("%s/temp/test.txt", tid))
	assert.NoError(t, err)
	// Iterate until done, the provider can devide into multiple parts (or not)
	var outTotal []byte
	for {
		out, moreData, err := reader.ReadOutput(ctx)
		assert.NoError(t, err)
		outTotal = append(outTotal, out...)
		if !moreData {
			break
		}
	}
	assert.Equal(t, "test123", string(outTotal))
	_, _, err = reader.ReadOutput(ctx)
	assert.Error(t, err)
	// Get Size (should be the number of bytes we stored)
	pSize, err = server.PathInfo(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(1), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(1), pSize.GetNumberOfFolders())
	// Get Size (should be the number of bytes we stored)
	pSize, err = server.PathInfo(ctx, fmt.Sprintf("%s/temp", tid))
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(1), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFolders())
	// Get Size (should be the number of bytes we stored)
	pSize, err = server.PathInfo(ctx, fmt.Sprintf("%s/temp/test.txt", tid))
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(1), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFolders())
	// Get Size (should be the number of bytes we stored)
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/newtest.txt", tid), []byte("anewtest"), true, fileTags)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/anothertest.txt", tid), []byte("newone"), true, fileTags)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/folder/anothertest.txt", tid), []byte("newone"), true, fileTags)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/folder/folder2/anothertest.txt", tid), []byte("newone"), true, fileTags)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	pSize, err = server.PathInfo(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, uint64(33), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(5), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(3), pSize.GetNumberOfFolders())
	// Get Size of non existing path (should be 0)
	pSize, err = server.PathInfo(ctx, fmt.Sprintf("%s/temp/test2.txt", tid))
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFolders())

	// Write another file, which end with an empty slice to close
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test2.txt", tid), []byte("test2"), true, nil)
	assert.NoError(t, err)
	assert.True(t, moreDataAllowed)
	// Close file
	moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test2.txt", tid), nil, false, nil)
	assert.NoError(t, err)
	assert.False(t, moreDataAllowed)
	// Check file
	info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test2.txt", tid))
	assert.NoError(t, err)
	assert.True(t, info.Exists)
	assert.True(t, info.IsLocked)
	assert.Equal(t, uint64(5), info.SizeInBytes)
	// Delete file
	_, err = server.DeletePath(ctx, &pbStorage.PathRequest{Path: fmt.Sprintf("%s/temp/test2.txt", tid)})
	assert.NoError(t, err)

	if testLargeFile {
		// Write an approx. 10MB file, which need to be read again
		for i := 0; i < 100; i++ {
			// 1k per line
			line := fmt.Sprintf("%01023d\n", i)
			// 100k per block
			block := ""
			for j := 0; j < 100; j++ {
				block += line
			}
			moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test3.txt", tid), []byte(block), true, nil)
			assert.NoError(t, err)
			assert.True(t, moreDataAllowed)
		}
		// Close file
		moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test3.txt", tid), nil, false, nil)
		assert.NoError(t, err)
		assert.False(t, moreDataAllowed)
		// Check file
		info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test3.txt", tid))
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.True(t, info.IsLocked)
		assert.Equal(t, uint64(1024*100*100), info.SizeInBytes)
		// Read file
		reader, err = server.createReader(ctx, fmt.Sprintf("%s/temp/test3.txt", tid))
		assert.NoError(t, err)
		// Iterate until done, the provider can devide into multiple parts (or not)
		totalBytes := 0
		for {
			out, moreData, err := reader.ReadOutput(ctx)
			assert.NoError(t, err)
			totalBytes += len(out)
			if !moreData {
				break
			}
		}
		assert.Equal(t, 1024*100*100, totalBytes)
		// Delete file
		_, err = server.DeletePath(ctx, &pbStorage.PathRequest{Path: fmt.Sprintf("%s/temp/test3.txt", tid)})
		assert.NoError(t, err)
	}
	// Test with Writer mode
	{
		// Test CreateWriter
		writer, err := server.createWriter(ctx, fmt.Sprintf("%s/temp/test4.txt", tid), fileTags)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		// Write some data
		moreAllowed, err := writer.WriteInput(ctx, []byte("test4a"), true)
		assert.NoError(t, err)
		assert.True(t, moreAllowed)
		// Write more data
		moreAllowed, err = writer.WriteInput(ctx, []byte("test4bb"), true)
		assert.NoError(t, err)
		assert.True(t, moreAllowed)
		// Wait a few seconds before checking, the write is async
		time.Sleep(2 * time.Second)
		// Check file
		info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test4.txt", tid))
		assert.NoError(t, err)
		assert.False(t, info.Exists)
		// Write more data
		moreAllowed, err = writer.WriteInput(ctx, []byte("test4ccc"), false)
		assert.NoError(t, err)
		assert.False(t, moreAllowed)
		_, err = writer.WriteInput(ctx, []byte("test4dddd"), false)
		assert.Error(t, err)
		// Wait a few seconds before checking, the close is async
		time.Sleep(2 * time.Second)
		// Check file
		info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test4.txt", tid))
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.True(t, info.IsLocked)
		assert.Equal(t, uint64(len("test4atest4bbtest4ccc")), info.SizeInBytes)
		assert.Len(t, info.Tags, 1)
		assert.Equal(t, "ftestv", *info.Tags.GetValue("ftestk"))
		// Get Data
		reader, err = server.createReader(ctx, fmt.Sprintf("%s/temp/test4.txt", tid))
		assert.NoError(t, err)
		// Iterate until done, the provider can devide into multiple parts (or not)
		outTotal = []byte{}
		for {
			out, moreData, err := reader.ReadOutput(ctx)
			assert.NoError(t, err)
			outTotal = append(outTotal, out...)
			if !moreData {
				break
			}
		}
		assert.Equal(t, "test4atest4bbtest4ccc", string(outTotal))
		moreDataAllowed, err = server.write(ctx, fmt.Sprintf("%s/temp/test5.txt", tid), []byte("test4atest4bb"), true, nil)
		assert.NoError(t, err)
		assert.True(t, moreDataAllowed)
		info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test5.txt", tid))
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.False(t, info.IsLocked)
		assert.Equal(t, uint64(len("test4atest4bb")), info.SizeInBytes)
		// Create a writer on the existing data
		writer, err = server.createWriter(ctx, fmt.Sprintf("%s/temp/test5.txt", tid), nil)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		// Write more data
		moreAllowed, err = writer.WriteInput(ctx, []byte("test4ccc"), false)
		assert.NoError(t, err)
		assert.False(t, moreAllowed)
		_, err = writer.WriteInput(ctx, []byte("test4dddd"), false)
		assert.Error(t, err)
		// Check file
		info, err = server.getObjectInfo(ctx, fmt.Sprintf("%s/temp/test4.txt", tid))
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.True(t, info.IsLocked)
		assert.Equal(t, uint64(len("test4atest4bbtest4ccc")), info.SizeInBytes)
		assert.Len(t, info.Tags, 1)
		assert.Equal(t, "ftestv", *info.Tags.GetValue("ftestk"))
	}
	// Delete Path
	_, err = server.DeletePath(ctx, &pbStorage.PathRequest{Path: tid})
	assert.NoError(t, err)
	// Get Size (should be 0 again)
	pSize, err = server.PathInfo(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), pSize.GetSizeInBytes())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFiles())
	assert.Equal(t, uint32(0), pSize.GetNumberOfFolders())
}
