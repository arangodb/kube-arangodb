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

package v1

import (
	"context"
	"fmt"
	"io"
	"path"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (i *implementation) PickUpJob(ctx context.Context, _ *pbSharedV1.Empty) (*pbLinkV1.PickUpJobResponse, error) {
	job, err := i.store.PickUp(ctx)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return &pbLinkV1.PickUpJobResponse{}, nil
	}

	return &pbLinkV1.PickUpJobResponse{
		Id: util.NewType(job.Id),
	}, nil
}

func (i *implementation) GetJob(ctx context.Context, req *pbLinkV1.GetJobRequest) (*pbLinkV1.Job, error) {
	job, _, err := i.store.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (i *implementation) UpdateJobStatus(ctx context.Context, req *pbLinkV1.UpdateJobStatusRequest) (*pbLinkV1.UpdateJobStatusResponse, error) {
	job, err := i.store.UpdateStatus(ctx, req.GetId(), req.GetStatus())
	if err != nil {
		return nil, err
	}

	return &pbLinkV1.UpdateJobStatusResponse{
		Job: job,
	}, nil
}

func (i *implementation) UploadFile(ctx context.Context, req *pbLinkV1.UploadFileRequest) (*pbLinkV1.UploadFileResponse, error) {
	if req.GetJobId() == "" || req.GetName() == "" {
		return nil, fmt.Errorf("job_id and name are required")
	}

	filePath := path.Join(FileStorePath(i.linkID, req.GetJobId()), req.GetName())

	writer, err := i.storage.WriteObject(ctx)
	if err != nil {
		return nil, err
	}

	if err := writer.Send(&pbStorageV2.StorageV2WriteObjectRequest{
		Path:  &pbStorageV2.StorageV2Path{Path: filePath},
		Chunk: req.GetData(),
	}); err != nil {
		return nil, err
	}

	resp, err := writer.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return &pbLinkV1.UploadFileResponse{
		Bytes:    resp.GetBytes(),
		Checksum: resp.GetChecksum(),
	}, nil
}

func (i *implementation) BatchUploadFiles(stream pbLinkV1.LinkV1Internal_BatchUploadFilesServer) error {
	var results []*pbLinkV1.UploadFileResponse
	var currentWriter pbStorageV2.StorageV2_WriteObjectClient
	var currentName string

	finishCurrent := func() error {
		if currentWriter == nil {
			return nil
		}

		resp, err := currentWriter.CloseAndRecv()
		if err != nil {
			return err
		}

		results = append(results, &pbLinkV1.UploadFileResponse{
			Bytes:    resp.GetBytes(),
			Checksum: resp.GetChecksum(),
		})
		currentWriter = nil
		currentName = ""
		return nil
	}

	for {
		req, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return recvErr
		}

		if name := req.GetName(); name != "" && name != currentName {
			if err := finishCurrent(); err != nil {
				return err
			}

			jobID := req.GetJobId()
			if jobID == "" {
				return fmt.Errorf("job_id must be set on the first chunk of each file")
			}

			filePath := path.Join(FileStorePath(i.linkID, jobID), name)
			var err error
			currentWriter, err = i.storage.WriteObject(stream.Context())
			if err != nil {
				return err
			}
			currentName = name

			if err := currentWriter.Send(&pbStorageV2.StorageV2WriteObjectRequest{
				Path:  &pbStorageV2.StorageV2Path{Path: filePath},
				Chunk: req.GetChunk(),
			}); err != nil {
				return err
			}
			continue
		}

		if currentWriter == nil {
			return fmt.Errorf("name must be set on the first chunk")
		}

		if err := currentWriter.Send(&pbStorageV2.StorageV2WriteObjectRequest{
			Chunk: req.GetChunk(),
		}); err != nil {
			return err
		}
	}

	if err := finishCurrent(); err != nil {
		return err
	}

	return stream.SendAndClose(&pbLinkV1.BatchUploadFilesResponse{
		Files: results,
	})
}

func (i *implementation) UpdateInfo(ctx context.Context, info *pbLinkV1.LinkInfo) (*pbSharedV1.Empty, error) {
	i.info = info
	return &pbSharedV1.Empty{}, nil
}
