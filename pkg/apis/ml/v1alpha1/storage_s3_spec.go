//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	"github.com/pkg/errors"
)

type ArangoMLStorageS3Spec struct {
	// Endpoint specifies the S3 API-compatible endpoint which implements storage
	// Required
	Endpoint string `json:"endpoint"`
	// DisableSSL if set to true, no certificate checks will be performed for Endpoint
	// +doc/default: false
	DisableSSL bool `json:"disableSSL,omitempty"`
	// Region defines the availability zone name. If empty, defaults to 'us-east-1'
	// +doc/default: ""
	Region string `json:"region,omitempty"`
	// BucketName specifies the name of the bucket
	// Required
	BucketName string `json:"bucketName"`
	// CredentialsSecretName specifies the name of the secret containing AccessKey and SecretKey for S3 API authorization
	// Required
	CredentialsSecretName string `json:"credentialsSecret"`
}

func (s *ArangoMLStorageS3Spec) Validate() error {
	if s.BucketName == "" {
		return errors.New("S3 BucketName must be not empty")
	}

	if s.Endpoint == "" {
		return errors.New("S3 Endpoint must be not empty")
	}

	if s.CredentialsSecretName == "" {
		return errors.New("S3 CredentialsSecretName must be not empty")
	}
	return nil
}
