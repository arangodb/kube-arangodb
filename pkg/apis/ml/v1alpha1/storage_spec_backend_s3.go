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
	"net/url"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLStorageSpecBackendS3 struct {
	// Endpoint specifies the S3 API-compatible endpoint which implements storage
	// Required
	Endpoint *string `json:"endpoint"`
	// BucketName specifies the name of the bucket
	// Required
	BucketName *string `json:"bucketName"`
	// CredentialsSecretName specifies the name of the secret containing AccessKey and SecretKey for S3 API authorization
	// Required
	CredentialsSecretName *string `json:"credentialsSecretName"`
	// AllowInsecure if set to true, the Endpoint certificates won't be checked
	// +doc/default: false
	AllowInsecure *bool `json:"allowInsecure,omitempty"`
	// CASecretName if not empty, the given secret will be used to check the authenticity of Endpoint
	// The specified `Secret`, must contain the following data fields:
	// - `ca.crt` PEM encoded public key of the CA certificate
	// - `ca.key` PEM encoded private key of the CA certificate
	// +doc/default: ""
	CASecretName *string `json:"caSecretName,omitempty"`
	// Region defines the availability zone name. If empty, defaults to 'us-east-1'
	// +doc/default: ""
	Region *string `json:"region,omitempty"`
}

func (s *ArangoMLStorageSpecBackendS3) Validate() error {
	if s == nil {
		s = &ArangoMLStorageSpecBackendS3{}
	}

	if s.GetBucketName() == "" {
		return errors.New("bucketName must be not empty")
	}

	if s.GetEndpoint() == "" {
		return errors.New("endpoint must be not empty")
	}

	if _, err := url.Parse(s.GetEndpoint()); err != nil {
		return errors.Newf("invalid endpoint URL was provided: %s", err.Error())
	}

	if s.GetCredentialsSecretName() == "" {
		return errors.New("credentialsSecretName must be not empty")
	}
	return nil
}

func (s *ArangoMLStorageSpecBackendS3) GetEndpoint() string {
	if s.Endpoint == nil {
		return ""
	}
	return *s.Endpoint
}

func (s *ArangoMLStorageSpecBackendS3) GetBucketName() string {
	if s.BucketName == nil {
		return ""
	}
	return *s.BucketName
}

func (s *ArangoMLStorageSpecBackendS3) GetCredentialsSecretName() string {
	if s.CredentialsSecretName == nil {
		return ""
	}
	return *s.CredentialsSecretName
}

func (s *ArangoMLStorageSpecBackendS3) GetAllowInsecure() bool {
	if s.AllowInsecure == nil {
		return false
	}
	return *s.AllowInsecure
}

func (s *ArangoMLStorageSpecBackendS3) GetCASecretName() string {
	if s.CASecretName == nil {
		return ""
	}
	return *s.CASecretName
}

func (s *ArangoMLStorageSpecBackendS3) GetRegion() string {
	if s.Region == nil {
		return ""
	}
	return *s.Region
}
