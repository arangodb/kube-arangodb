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

package v1beta1

import (
	"net/url"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformStorageSpecBackendAzureBlobStorage struct {
	// TenantID specifies the Azure TenantID
	// +doc/required
	TenantID *string `json:"tenantID,omitempty"`

	// AccountName specifies the Azure Storage AccountName
	// used in format https://<account>.blob.core.windows.net/
	// +doc/required
	AccountName *string `json:"accountName,omitempty"`

	// Endpoint specifies the Azure Storage custom endpoint
	// +doc/required
	Endpoint *string `json:"endpoint,omitempty"`

	// BucketName specifies the name of the bucket
	// +doc/required
	BucketName *string `json:"bucketName,omitempty"`

	// BucketPath specifies the Prefix within the bucket
	// +doc/default:
	BucketPrefix *string `json:"bucketPath,omitempty"`

	// CredentialsSecret specifies the Kubernetes Secret containing ClientID and ClientSecret for Azure API authorization
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	CredentialsSecret *sharedApi.Object `json:"credentialsSecret"`
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpecBackendAzureBlobStorage{}
	}

	var errs []error

	if end := s.GetEndpoint(); end != "" {
		if _, err := url.Parse(s.GetEndpoint()); err != nil {
			errs = append(errs, shared.PrefixResourceErrors("endpoint", errors.Errorf("invalid URL: %s", err.Error())))
		}
	}
	if acc := s.GetTenantID(); acc == "" {
		errs = append(errs, shared.PrefixResourceErrors("tenantID", errors.Errorf("TenantID needs to be defined")))
	}

	if acc := s.GetAccountName(); acc == "" && s.GetEndpoint() == "" {
		errs = append(errs, shared.PrefixResourceErrors("accountName", errors.Errorf("AccountName needs to be defined")))
	}

	errs = append(errs,
		shared.PrefixResourceErrors("credentialsSecret", s.GetCredentialsSecret().Validate()),
		shared.PrefixResourceError("bucketName", shared.ValidateRequired(s.BucketName, shared.ValidateResourceName)),
	)

	return shared.WithErrors(errs...)
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetAccountName() string {
	if s == nil || s.AccountName == nil {
		return ""
	}
	return *s.AccountName
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetEndpoint() string {
	if s == nil || s.Endpoint == nil {
		return ""
	}
	return *s.Endpoint
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetTenantID() string {
	if s == nil || s.TenantID == nil {
		return ""
	}
	return *s.TenantID
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetBucketName() string {
	if s == nil || s.BucketName == nil {
		return ""
	}
	return *s.BucketName
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetBucketPrefix() string {
	if s == nil || s.BucketPrefix == nil {
		return ""
	}
	return *s.BucketPrefix
}

func (s *ArangoPlatformStorageSpecBackendAzureBlobStorage) GetCredentialsSecret() *sharedApi.Object {
	if s == nil || s.CredentialsSecret == nil {
		return &sharedApi.Object{}
	}
	return s.CredentialsSecret
}
