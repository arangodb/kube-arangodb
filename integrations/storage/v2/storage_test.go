//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func listAllFilesHelper(t *testing.T, ctx context.Context, h pbStorageV2.StorageV2Client, prefix string) []*pbStorageV2.StorageV2Object {
	var r []*pbStorageV2.StorageV2Object

	res, err := h.ListObjects(ctx, &pbStorageV2.StorageV2ListObjectsRequest{
		Path: &pbStorageV2.StorageV2Path{
			Path: prefix,
		},
	})
	require.NoError(t, err)

	for {
		files, err := res.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		r = append(r, files.GetFiles()...)
	}

	return r
}

func Test_List(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	h := Client(t, ctx, func(c Configuration) Configuration {
		c.S3.MaxListKeys = util.NewType[int64](32)
		return c
	})
	testFileListing(t, ctx, h)
}

func Test_Flow_16(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	h := Client(t, ctx, func(c Configuration) Configuration {
		c.S3.MaxListKeys = util.NewType[int64](32)
		return c
	})

	testS3BucketFileHandling(t, ctx, h, 16)
}

func Test_Flow_1024(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	h := Client(t, ctx, func(c Configuration) Configuration {
		c.S3.MaxListKeys = util.NewType[int64](32)
		return c
	})

	testS3BucketFileHandling(t, ctx, h, 1024)
}

func Test_Flow_1048576(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	h := Client(t, ctx, func(c Configuration) Configuration {
		c.S3.MaxListKeys = util.NewType[int64](32)
		return c
	})

	testS3BucketFileHandling(t, ctx, h, 1024*1024)
}

func Test_Flow_4194304(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	h := Client(t, ctx, func(c Configuration) Configuration {
		c.S3.MaxListKeys = util.NewType[int64](32)
		return c
	})

	testS3BucketFileHandling(t, ctx, h, 4*1024*1024)
}

func testFileListing(t *testing.T, ctx context.Context, h pbStorageV2.StorageV2Client) {
	prefix := fmt.Sprintf("%s/", uuid.NewUUID())
	t.Run("List", func(t *testing.T) {
		var files []string

		t.Run("RenderFileNames", func(t *testing.T) {
			for i := 0; i < 128; i++ {
				files = append(files, fmt.Sprintf("%sfile%04d", prefix, i))
				files = append(files, fmt.Sprintf("%spath%04d/file", prefix, i))
			}
		})

		sort.Strings(files)

		t.Logf("Files: %d", len(files))

		data := make([]byte, 1024)
		n, err := rand.Read(data)
		require.NoError(t, err)
		require.EqualValues(t, 1024, n)

		checksum := util.SHA256(data)

		t.Run("UploadAll", func(t *testing.T) {
			util.ParallelProcess(func(in string) {
				wr, err := h.WriteObject(ctx)
				require.NoError(t, err)

				buff := make([]byte, 1024)

				cf := bytes.NewReader(data)

				for {
					n, err := cf.Read(buff)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}

						require.NoError(t, err)
					}

					require.NoError(t, wr.Send(&pbStorageV2.StorageV2WriteObjectRequest{
						Path: &pbStorageV2.StorageV2Path{
							Path: in,
						},
						Chunk: buff[:n],
					}))
				}

				ds, err := wr.CloseAndRecv()
				require.NoError(t, err)

				require.NotNil(t, ds)
				require.EqualValues(t, checksum, ds.GetChecksum())
				require.EqualValues(t, len(data), ds.GetBytes())
			}, 32, files)
		})

		t.Run("CheckAll", func(t *testing.T) {
			util.ParallelProcess(func(in string) {
				wr, err := h.HeadObject(ctx, &pbStorageV2.StorageV2HeadObjectRequest{
					Path: &pbStorageV2.StorageV2Path{Path: in},
				})
				require.NoError(t, err)
				require.NotNil(t, wr)

				require.EqualValues(t, len(data), wr.GetInfo().GetSize())
			}, 32, files)
		})

		t.Run("List", func(t *testing.T) {
			revcFiles := listAllFilesHelper(t, ctx, h, prefix)

			require.Len(t, revcFiles, len(files))

			for id := range files {
				require.EqualValues(t, strings.TrimPrefix(files[id], prefix), revcFiles[id].GetPath().GetPath())
				require.EqualValues(t, revcFiles[id].GetInfo().GetSize(), len(data))
			}
		})

		t.Run("ListSubFolder", func(t *testing.T) {
			revcFiles := listAllFilesHelper(t, ctx, h, fmt.Sprintf("%spath0000/", prefix))

			require.Len(t, revcFiles, 1)

			require.EqualValues(t, "file", revcFiles[0].GetPath().GetPath())
			require.EqualValues(t, len(data), revcFiles[0].GetInfo().GetSize())
		})

		t.Run("ListMisSubFolder", func(t *testing.T) {
			revcFiles := listAllFilesHelper(t, ctx, h, fmt.Sprintf("%snon-existent/", prefix))

			require.Len(t, revcFiles, 0)
		})

		t.Run("DeleteAll", func(t *testing.T) {
			util.ParallelProcess(func(in string) {
				wr, err := h.DeleteObject(ctx, &pbStorageV2.StorageV2DeleteObjectRequest{
					Path: &pbStorageV2.StorageV2Path{Path: in},
				})
				require.NoError(t, err)
				require.NotNil(t, wr)
			}, 32, files)
		})
	})
}

func testS3BucketFileHandling(t *testing.T, ctx context.Context, h pbStorageV2.StorageV2Client, size int) {
	t.Run(fmt.Sprintf("Size:%d", size), func(t *testing.T) {
		prefix := fmt.Sprintf("%s/", uuid.NewUUID())
		name := fmt.Sprintf("%stest.local", prefix)
		nameTwo := fmt.Sprintf("%stest.local.two", prefix)
		t.Logf("File: %s", name)

		dataOne := make([]byte, size)

		n, err := rand.Read(dataOne)
		require.NoError(t, err)
		require.EqualValues(t, size, n)

		checksumOne := util.SHA256(dataOne)

		dataTwo := make([]byte, size)

		n, err = rand.Read(dataTwo)
		require.NoError(t, err)
		require.EqualValues(t, size, n)

		checksumTwo := util.SHA256(dataTwo)

		t.Logf("Checksum One: %s", checksumOne)
		t.Logf("Checksum Two: %s", checksumTwo)

		require.NotEqual(t, checksumTwo, checksumOne)

		t.Run("Check if object exists", func(t *testing.T) {
			resp, err := h.HeadObject(ctx, &pbStorageV2.StorageV2HeadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})

			require.EqualValues(t, codes.NotFound, errors.GRPCCode(err))
			require.Nil(t, resp)
		})

		t.Run("Send Object", func(t *testing.T) {
			wr, err := h.WriteObject(ctx)
			require.NoError(t, err)

			buff := make([]byte, 1024)

			cf := bytes.NewReader(dataOne)

			for {
				n, err := cf.Read(buff)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					require.NoError(t, err)
				}

				require.NoError(t, wr.Send(&pbStorageV2.StorageV2WriteObjectRequest{
					Path: &pbStorageV2.StorageV2Path{
						Path: name,
					},
					Chunk: buff[:n],
				}))
			}

			ds, err := wr.CloseAndRecv()
			require.NoError(t, err)

			require.NotNil(t, ds)
			require.EqualValues(t, checksumOne, ds.GetChecksum())
			require.EqualValues(t, len(dataOne), ds.GetBytes())
		})

		t.Run("Re-Check if object exists", func(t *testing.T) {
			resp, err := h.HeadObject(ctx, &pbStorageV2.StorageV2HeadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})

			require.EqualValues(t, codes.OK, errors.GRPCCode(err))
			require.NotNil(t, resp)

			require.EqualValues(t, len(dataOne), resp.GetInfo().GetSize())
		})

		t.Run("Download Object", func(t *testing.T) {
			wr, err := h.ReadObject(ctx, &pbStorageV2.StorageV2ReadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})
			require.NoError(t, err)

			data := bytes.NewBuffer(nil)

			for {
				resp, err := wr.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)

				_, err = util.WriteAll(data, resp.GetChunk())
				require.NoError(t, err)
			}

			pdata := data.Bytes()

			require.Len(t, pdata, size)

			pchecksum := util.SHA256(pdata)
			require.EqualValues(t, checksumOne, pchecksum)
		})

		t.Run("Re-Send Object", func(t *testing.T) {
			wr, err := h.WriteObject(ctx)
			require.NoError(t, err)

			buff := make([]byte, 1024)

			cf := bytes.NewReader(dataTwo)

			for {
				n, err := cf.Read(buff)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					require.NoError(t, err)
				}

				require.NoError(t, wr.Send(&pbStorageV2.StorageV2WriteObjectRequest{
					Path: &pbStorageV2.StorageV2Path{
						Path: name,
					},
					Chunk: buff[:n],
				}))
			}

			ds, err := wr.CloseAndRecv()
			require.NoError(t, err)

			require.NotNil(t, ds)
			require.EqualValues(t, checksumTwo, ds.GetChecksum())
			require.EqualValues(t, len(dataTwo), ds.GetBytes())
		})

		t.Run("List Objects", func(t *testing.T) {
			revcFiles := listAllFilesHelper(t, ctx, h, prefix)

			t.Logf("Size: %d", len(revcFiles))
		})

		t.Run("Send Second Object", func(t *testing.T) {
			wr, err := h.WriteObject(ctx)
			require.NoError(t, err)

			buff := make([]byte, 1024)

			cf := bytes.NewReader(dataOne)

			for {
				n, err := cf.Read(buff)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					require.NoError(t, err)
				}

				require.NoError(t, wr.Send(&pbStorageV2.StorageV2WriteObjectRequest{
					Path: &pbStorageV2.StorageV2Path{
						Path: nameTwo,
					},
					Chunk: buff[:n],
				}))
			}

			ds, err := wr.CloseAndRecv()
			require.NoError(t, err)

			require.NotNil(t, ds)
			require.EqualValues(t, checksumOne, ds.GetChecksum())
			require.EqualValues(t, len(dataOne), ds.GetBytes())
		})

		t.Run("Re-Download Object", func(t *testing.T) {
			wr, err := h.ReadObject(ctx, &pbStorageV2.StorageV2ReadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})
			require.NoError(t, err)

			data := bytes.NewBuffer(nil)

			for {
				resp, err := wr.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)

				_, err = util.WriteAll(data, resp.GetChunk())
				require.NoError(t, err)
			}

			pdata := data.Bytes()

			require.Len(t, pdata, size)

			pchecksum := util.SHA256(pdata)
			require.EqualValues(t, checksumTwo, pchecksum)
		})

		t.Run("Delete Object", func(t *testing.T) {
			wr, err := h.DeleteObject(ctx, &pbStorageV2.StorageV2DeleteObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})
			require.NoError(t, err)
			require.NotNil(t, wr)
		})

		t.Run("Delete Second Object", func(t *testing.T) {
			wr, err := h.DeleteObject(ctx, &pbStorageV2.StorageV2DeleteObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: nameTwo},
			})
			require.NoError(t, err)
			require.NotNil(t, wr)
		})

		t.Run("Re-Check if deleted object exists", func(t *testing.T) {
			resp, err := h.HeadObject(ctx, &pbStorageV2.StorageV2HeadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})

			require.EqualValues(t, codes.NotFound, errors.GRPCCode(err))
			require.Nil(t, resp)
		})

		t.Run("Download Deleted Object", func(t *testing.T) {
			wr, err := h.ReadObject(ctx, &pbStorageV2.StorageV2ReadObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})
			require.NoError(t, err)

			resp, err := wr.Recv()
			require.EqualValues(t, codes.NotFound, errors.GRPCCode(err))
			require.Nil(t, resp)
		})

		t.Run("Delete Deleted Object", func(t *testing.T) {
			wr, err := h.DeleteObject(ctx, &pbStorageV2.StorageV2DeleteObjectRequest{
				Path: &pbStorageV2.StorageV2Path{Path: name},
			})
			require.NoError(t, err)
			require.NotNil(t, wr)
		})
	})
}
