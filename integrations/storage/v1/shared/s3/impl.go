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
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	pbStorage "github.com/arangodb-managed/integration-apis/bucket-service/v1"

	pbStorageV1 "github.com/arangodb/kube-arangodb/integrations/storage/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

const (
	// LockedKey contains the key for indication if an object is locked
	LockedKey = "locked"
	// LockedValue contains the value for indicating it's locked
	LockedValue = "true"
)

var _ ShutdownableBucketServiceServer = &s3impl{}

type ShutdownableBucketServiceServer interface {
	pbStorage.BucketServiceServer

	svc.Handler

	svc.Shutdown
}

type s3impl struct {
	bucket string
	region string

	registeredWritersMutex sync.Mutex
	registeredWriters      []*InputWriter

	client *s3.S3

	pbStorage.UnimplementedBucketServiceServer
}

func (s *s3impl) Name() string {
	return pbStorageV1.Name
}

func (s *s3impl) Health() svc.HealthState {
	return svc.Healthy
}

func (s *s3impl) Register(registrar *grpc.Server) {
	pbStorage.RegisterBucketServiceServer(registrar, s)
}

func NewS3Impl(cfg Configuration) (ShutdownableBucketServiceServer, error) {
	prov, err := cfg.Client.GetAWSSession()
	if err != nil {
		return nil, err
	}

	storageClient := s3.New(prov, aws.NewConfig().WithRegion(cfg.Client.GetRegion()))

	return &s3impl{
		client: storageClient,
		bucket: cfg.BucketName,
		region: cfg.Client.GetRegion(),
	}, nil
}

func (s *s3impl) getPrefix(path string) *string {
	query := ""
	if path != "." {
		query = path
	}
	return util.NewType(query)
}

func (s *s3impl) getRepositoryURL(path string) *pbStorage.RepositoryURL {
	prefix := s.getPrefix(path)
	repositoryBucketPath := fmt.Sprintf("%s/%s", s.bucket, *prefix)
	repositoryURL := fmt.Sprintf("s3:%s", repositoryBucketPath)
	return &pbStorage.RepositoryURL{
		Url:        repositoryURL,
		BucketPath: repositoryBucketPath,
	}
}

// Shutdown is called when the service needs to shutdown
func (s *s3impl) Shutdown(cancel context.CancelFunc) {
	logger.Debug("Shutdown received")

	// Lock the registered writers access
	s.registeredWritersMutex.Lock()
	defer s.registeredWritersMutex.Unlock()

	// Close all writers (async)
	g, _ := errgroup.WithContext(context.Background())
	for _, wp := range s.registeredWriters {
		w := *wp
		g.Go(func() error {
			logger.Debug("Close writer")
			if err := w.Close(); err != nil {
				logger.Err(err).Debug("Close writer failed")
				// Ignore
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Err(err).Debug("Wait failed")
		// Continue
	}
	// We are done, cancel context
	logger.Debug("Service completely done, cancel context")
	cancel()
}
