//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"io"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbStorageV2.StorageV2Server = &implementation{}
var _ svc.Handler = &implementation{}

func New(cfg Configuration) (svc.Handler, error) {
	return newInternal(cfg)
}

func newInternal(c Configuration) (*implementation, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Invalid config")
	}

	io, err := c.IO(shutdown.Context())
	if err != nil {
		return nil, err
	}

	return &implementation{
		io: io,
	}, nil
}

type implementation struct {
	io pbImplStorageV2Shared.IO

	pbStorageV2.UnimplementedStorageV2Server
}

func (i *implementation) Name() string {
	return pbStorageV2.Name
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbStorageV2.RegisterStorageV2Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return nil
}

func (i *implementation) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (i *implementation) WriteObject(server pbStorageV2.StorageV2_WriteObjectServer) error {
	ctx, c := context.WithCancel(server.Context())
	defer c()

	log := logger.Str("func", "WriteObject")

	msg, err := server.Recv()
	if err == io.EOF || errors.IsGRPCCode(err, codes.Canceled) {
		return io.ErrUnexpectedEOF
	}

	path := msg.GetPath().GetPath()
	if path == "" {
		log.Debug("path missing")
		return status.Error(codes.InvalidArgument, "path missing")
	}

	wd, err := i.io.Write(ctx, path)
	if err != nil {
		return err
	}

	if _, err := util.WriteAll(wd, msg.GetChunk()); err != nil {
		return err
	}

	for {
		msg, err := server.Recv()
		if errors.IsGRPCCode(err, codes.Canceled) {
			c()
			return io.ErrUnexpectedEOF
		}

		if errors.Is(err, io.EOF) {
			checksum, bytes, err := wd.Close(ctx)
			if err != nil {
				return err
			}

			if err := server.SendAndClose(&pbStorageV2.StorageV2WriteObjectResponse{
				Bytes:    bytes,
				Checksum: checksum,
			}); err != nil {
				log.Err(err).Debug("Failed to send WriteObjectControl message")
				return err
			}

			return nil
		}

		if err != nil {
			return err
		}

		if msg.GetPath() != nil {
			if path != msg.GetPath().GetPath() {
				log.Debug("path changed")
				return status.Error(codes.InvalidArgument, "path changed")
			}
		}

		if _, err := util.WriteAll(wd, msg.GetChunk()); err != nil {
			return err
		}
	}
}

func (i *implementation) ReadObject(req *pbStorageV2.StorageV2ReadObjectRequest, server pbStorageV2.StorageV2_ReadObjectServer) error {
	log := logger.Str("func", "ReadObject").Str("path", req.GetPath().GetPath())
	ctx := server.Context()
	path := req.GetPath().GetPath()
	if path == "" {
		return status.Errorf(codes.InvalidArgument, "path missing")
	}

	rd, err := i.io.Read(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status.Errorf(codes.NotFound, "file not found")
		}

		return err
	}

	buff := pbImplStorageV2Shared.NewBuffer(pbImplStorageV2Shared.MaxChunkBytes)

	for {
		n, err := rd.Read(buff)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			if errors.Is(err, os.ErrNotExist) {
				return status.Errorf(codes.NotFound, "file not found")
			}

			return err
		}

		// Send chunk to caller
		if err := server.Send(&pbStorageV2.StorageV2ReadObjectResponse{
			Chunk: buff[:n],
		}); err != nil {
			log.Err(err).Debug("Failed to send ReadObjectChunk")
			return err
		}
	}
}

func (i *implementation) HeadObject(ctx context.Context, req *pbStorageV2.StorageV2HeadObjectRequest) (*pbStorageV2.StorageV2HeadObjectResponse, error) {
	log := logger.Str("func", "HeadObject").Str("path", req.GetPath().GetPath())

	// Check request fields
	path := req.GetPath().GetPath()
	if path == "" {
		return nil, status.Error(codes.InvalidArgument, "path missing")
	}

	info, err := i.io.Head(ctx, path)
	if err != nil {
		log.Err(err).Debug("getObjectInfo failed")
		return nil, err
	}
	if info == nil {
		return nil, status.Error(codes.NotFound, path)
	}

	return &pbStorageV2.StorageV2HeadObjectResponse{
		Info: &pbStorageV2.StorageV2ObjectInfo{
			Size:        info.Size,
			LastUpdated: timestamppb.New(info.LastUpdatedAt),
		},
	}, nil
}

func (i *implementation) DeleteObject(ctx context.Context, req *pbStorageV2.StorageV2DeleteObjectRequest) (*pbStorageV2.StorageV2DeleteObjectResponse, error) {
	log := logger.Str("func", "DeleteObject").Str("path", req.GetPath().GetPath())

	// Check request fields
	path := req.GetPath().GetPath()
	if path == "" {
		return nil, status.Error(codes.InvalidArgument, "path missing")
	}

	deleted, err := i.io.Delete(ctx, path)
	if err != nil {
		log.Err(err).Debug("deleteObject failed")
		return nil, err
	}

	if deleted {
		return &pbStorageV2.StorageV2DeleteObjectResponse{}, nil
	}

	return nil, status.Error(codes.NotFound, "Object Not Found")
}

func (i *implementation) ListObjects(req *pbStorageV2.StorageV2ListObjectsRequest, server pbStorageV2.StorageV2_ListObjectsServer) error {
	log := logger.Str("func", "ReadObject").Str("path", req.GetPath().GetPath())
	ctx := server.Context()
	path := req.GetPath().GetPath()

	lister, err := i.io.List(ctx, path)
	if err != nil {
		log.Err(err).Debug("listObjects failed")
		return err
	}

	for {
		files, err := lister.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			log.Err(err).Debug("listObjects failed")
			return err
		}

		ret := make([]*pbStorageV2.StorageV2Object, len(files))

		for id := range files {
			ret[id] = &pbStorageV2.StorageV2Object{
				Path: &pbStorageV2.StorageV2Path{
					Path: files[id].Key,
				},
				Info: &pbStorageV2.StorageV2ObjectInfo{
					Size:        files[id].Info.Size,
					LastUpdated: timestamppb.New(files[id].Info.LastUpdatedAt),
				},
			}
		}

		if err := server.Send(&pbStorageV2.StorageV2ListObjectsResponse{
			Files: ret,
		}); err != nil {
			log.Err(err).Debug("listObjects failed")
			return err
		}
	}
}

func (i *implementation) Init(ctx context.Context, in *pbStorageV2.StorageV2InitRequest) (*pbStorageV2.StorageV2InitResponse, error) {
	if err := i.io.Init(ctx, &pbImplStorageV2Shared.InitOptions{
		Create: util.NewPointer(in.Create),
	}); err != nil {
		return nil, err
	}

	return &pbStorageV2.StorageV2InitResponse{}, nil
}
