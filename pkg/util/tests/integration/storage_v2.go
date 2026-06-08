//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	goStrings "strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
)

// NewStorageV2Client returns a filesystem-backed implementation of StorageV2Client for testing.
// Creates three directories under t.TempDir():
//   - files/     — current files (seed data here before tests)
//   - uploads/   — files being written (in-progress uploads)
//   - downloads/ — files copied from files/ on read
//
// On WriteObject CloseAndRecv, the upload is moved from uploads/ to files/.
// On ReadObject, the file is copied from files/ to downloads/ and served from there.
func NewStorageV2Client(t *testing.T) pbStorageV2.StorageV2Client {
	t.Helper()

	root := t.TempDir()

	filesDir := filepath.Join(root, "files")
	uploadsDir := filepath.Join(root, "uploads")
	downloadsDir := filepath.Join(root, "downloads")

	for _, d := range []string{filesDir, uploadsDir, downloadsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	return &fsStorageV2{
		filesDir:     filesDir,
		uploadsDir:   uploadsDir,
		downloadsDir: downloadsDir,
	}
}

type fsStorageV2 struct {
	filesDir     string
	uploadsDir   string
	downloadsDir string
}

func (s *fsStorageV2) Init(ctx context.Context, in *pbStorageV2.StorageV2InitRequest, opts ...grpc.CallOption) (*pbStorageV2.StorageV2InitResponse, error) {
	return &pbStorageV2.StorageV2InitResponse{}, nil
}

func (s *fsStorageV2) ReadObject(ctx context.Context, in *pbStorageV2.StorageV2ReadObjectRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbStorageV2.StorageV2ReadObjectResponse], error) {
	p := in.GetPath().GetPath()
	src := filepath.Join(s.filesDir, p)
	dst := filepath.Join(s.downloadsDir, p)

	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "Object %s not found", p)
		}
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, err
	}

	return &fsReadStream{data: data}, nil
}

func (s *fsStorageV2) WriteObject(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbStorageV2.StorageV2WriteObjectRequest, pbStorageV2.StorageV2WriteObjectResponse], error) {
	return &fsWriteStream{store: s}, nil
}

func (s *fsStorageV2) HeadObject(ctx context.Context, in *pbStorageV2.StorageV2HeadObjectRequest, opts ...grpc.CallOption) (*pbStorageV2.StorageV2HeadObjectResponse, error) {
	p := in.GetPath().GetPath()
	info, err := os.Stat(filepath.Join(s.filesDir, p))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "Object %s not found", p)
		}
		return nil, err
	}

	return &pbStorageV2.StorageV2HeadObjectResponse{
		Info: &pbStorageV2.StorageV2ObjectInfo{
			Size:        uint64(info.Size()),
			LastUpdated: timestamppb.New(info.ModTime()),
		},
	}, nil
}

func (s *fsStorageV2) DeleteObject(ctx context.Context, in *pbStorageV2.StorageV2DeleteObjectRequest, opts ...grpc.CallOption) (*pbStorageV2.StorageV2DeleteObjectResponse, error) {
	p := in.GetPath().GetPath()
	path := filepath.Join(s.filesDir, p)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, status.Errorf(codes.NotFound, "Object %s not found", p)
	}

	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return &pbStorageV2.StorageV2DeleteObjectResponse{}, nil
}

func (s *fsStorageV2) ListObjects(ctx context.Context, in *pbStorageV2.StorageV2ListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbStorageV2.StorageV2ListObjectsResponse], error) {
	prefix := in.GetPath().GetPath()
	var files []*pbStorageV2.StorageV2Object

	filepath.Walk(s.filesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(s.filesDir, path)
		if prefix == "" || goStrings.HasPrefix(rel, prefix) {
			files = append(files, &pbStorageV2.StorageV2Object{
				Path: &pbStorageV2.StorageV2Path{Path: rel},
				Info: &pbStorageV2.StorageV2ObjectInfo{
					Size:        uint64(info.Size()),
					LastUpdated: timestamppb.New(info.ModTime()),
				},
			})
		}
		return nil
	})

	return &fsListObjectsStream{files: files}, nil
}

// fsReadStream serves file data in a single chunk
type fsReadStream struct {
	grpc.ClientStream
	data []byte
	sent bool
}

func (s *fsReadStream) Recv() (*pbStorageV2.StorageV2ReadObjectResponse, error) {
	if s.sent {
		return nil, io.EOF
	}
	s.sent = true
	return &pbStorageV2.StorageV2ReadObjectResponse{Chunk: s.data}, nil
}

// fsWriteStream writes to uploads/, moves to files/ on close
type fsWriteStream struct {
	grpc.ClientStream
	store *fsStorageV2
	path  string
	buf   bytes.Buffer
}

func (s *fsWriteStream) Send(req *pbStorageV2.StorageV2WriteObjectRequest) error {
	if p := req.GetPath().GetPath(); p != "" {
		s.path = p
	}
	s.buf.Write(req.GetChunk())
	return nil
}

func (s *fsWriteStream) CloseAndRecv() (*pbStorageV2.StorageV2WriteObjectResponse, error) {
	data := s.buf.Bytes()

	// Write to uploads/
	uploadPath := filepath.Join(s.store.uploadsDir, s.path)
	if err := os.MkdirAll(filepath.Dir(uploadPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(uploadPath, data, 0o644); err != nil {
		return nil, err
	}

	// Move from uploads/ to files/
	filePath := filepath.Join(s.store.filesDir, s.path)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, err
	}
	if err := os.Rename(uploadPath, filePath); err != nil {
		return nil, err
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	return &pbStorageV2.StorageV2WriteObjectResponse{
		Bytes:    int64(len(data)),
		Checksum: checksum,
	}, nil
}

// fsListObjectsStream returns all files in a single chunk
type fsListObjectsStream struct {
	grpc.ClientStream
	files []*pbStorageV2.StorageV2Object
	sent  bool
}

func (s *fsListObjectsStream) Recv() (*pbStorageV2.StorageV2ListObjectsResponse, error) {
	if s.sent {
		return nil, io.EOF
	}
	s.sent = true
	return &pbStorageV2.StorageV2ListObjectsResponse{Files: s.files}, nil
}
