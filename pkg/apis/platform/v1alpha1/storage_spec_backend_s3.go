//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformStorageSpecBackendS3 struct {
	// BucketName specifies the name of the bucket
	// +doc/required
	BucketName *string `json:"bucketName,omitempty"`
	// BucketPath specifies the Prefix within the bucket
	// +doc/default:
	BucketPrefix *string `json:"bucketPath,omitempty"`
	// Endpoint specifies the S3 API-compatible endpoint which implements storage
	// +doc/required
	Endpoint *string `json:"endpoint"`
	// CredentialsSecret specifies the Kubernetes Secret containing AccessKey and SecretKey for S3 API authorization
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	CredentialsSecret *sharedApi.Object `json:"credentialsSecret"`
	// AllowInsecure if set to true, the Endpoint certificates won't be checked
	// +doc/default: false
	AllowInsecure *bool `json:"allowInsecure,omitempty"`
	// CASecret if not empty, the given Kubernetes Secret will be used to check the authenticity of Endpoint
	// The specified Secret, must contain the following data fields:
	// - `ca.crt` PEM encoded public key of the CA certificate
	// - `ca.key` PEM encoded private key of the CA certificate
	// +doc/default: nil
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	CASecret *sharedApi.Object `json:"caSecret,omitempty"`
	// Region defines the availability zone name.
	// +doc/default: ""
	Region *string `json:"region,omitempty"`
}

func (s *ArangoPlatformStorageSpecBackendS3) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpecBackendS3{}
	}

	var errs []error

	if s.GetEndpoint() == "" {
		errs = append(errs, shared.PrefixResourceErrors("endpoint", errors.New("must be not empty")))
	}

	if _, err := url.Parse(s.GetEndpoint()); err != nil {
		errs = append(errs, shared.PrefixResourceErrors("endpoint", errors.Errorf("invalid URL: %s", err.Error())))
	}

	errs = append(errs,
		shared.PrefixResourceErrors("credentialsSecret", s.GetCredentialsSecret().Validate()),
		shared.PrefixResourceError("bucket", shared.ValidateRequired(s.BucketName, shared.ValidateResourceName)),
	)

	if caSecret := s.GetCASecret(); !caSecret.IsEmpty() {
		errs = append(errs, shared.PrefixResourceErrors("caSecret", caSecret.Validate()))
	}

	return shared.WithErrors(errs...)
}

func (s *ArangoPlatformStorageSpecBackendS3) GetBucketName() string {
	if s == nil || s.BucketName == nil {
		return ""
	}
	return *s.BucketName
}

func (s *ArangoPlatformStorageSpecBackendS3) GetBucketPrefix() string {
	if s == nil || s.BucketPrefix == nil {
		return ""
	}
	return *s.BucketPrefix
}

func (s *ArangoPlatformStorageSpecBackendS3) GetEndpoint() string {
	if s == nil || s.Endpoint == nil {
		return ""
	}
	return *s.Endpoint
}

func (s *ArangoPlatformStorageSpecBackendS3) GetCredentialsSecret() *sharedApi.Object {
	if s == nil || s.CredentialsSecret == nil {
		return &sharedApi.Object{}
	}
	return s.CredentialsSecret
}

func (s *ArangoPlatformStorageSpecBackendS3) GetAllowInsecure() bool {
	if s == nil || s.AllowInsecure == nil {
		return false
	}
	return *s.AllowInsecure
}

func (s *ArangoPlatformStorageSpecBackendS3) GetCASecret() *sharedApi.Object {
	if s == nil || s.CASecret == nil {
		return &sharedApi.Object{}
	}
	return s.CASecret
}

func (s *ArangoPlatformStorageSpecBackendS3) GetRegion() string {
	if s == nil || s.Region == nil {
		return ""
	}
	return *s.Region
}
