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

package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
)

type Configuration struct {
	BucketName   string
	BucketPrefix string

	MaxListKeys *int64

	Client awsHelper.Config
}

func (c Configuration) New() (pbImplStorageV2Shared.IO, error) {
	prov, err := c.Client.GetAWSSession()
	if err != nil {
		return nil, err
	}

	storageClient := s3.New(prov, aws.NewConfig().WithRegion(c.Client.GetRegion()))

	return &ios{
		config:   c,
		client:   storageClient,
		uploader: s3manager.NewUploaderWithClient(storageClient),
		downloader: s3manager.NewDownloaderWithClient(storageClient, func(downloader *s3manager.Downloader) {
			downloader.Concurrency = 1
		}),
	}, nil
}
