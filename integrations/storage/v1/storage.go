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

package v1

import (
	pbImplStorageV1SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v1/shared/s3"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(cfg Configuration) (svc.Handler, error) {
	switch cfg.Type {
	case ConfigurationTypeS3:

		impl, err := pbImplStorageV1SharedS3.NewS3Impl(cfg.S3)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create S3 service server")
		}

		return impl, nil
	default:
		return nil, errors.New("currently only 's3' storage type is supported")
	}
}
