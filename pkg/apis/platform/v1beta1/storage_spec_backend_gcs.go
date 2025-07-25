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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformStorageSpecBackendGCS struct {
	// ProjectID specifies the GCP ProjectID
	// +doc/required
	ProjectID *string `json:"projectID,omitempty"`
	// BucketName specifies the name of the bucket
	// +doc/required
	BucketName *string `json:"bucketName,omitempty"`
	// BucketPath specifies the Prefix within the bucket
	// +doc/default:
	BucketPrefix *string `json:"bucketPath,omitempty"`
	// CredentialsSecret specifies the Kubernetes Secret containing Service Account JSON Key as data field
	// +doc/required
	// +doc/skip: namespace
	// +doc/skip: uid
	// +doc/skip: checksum
	CredentialsSecret *sharedApi.Object `json:"credentialsSecret"`
}

func (s *ArangoPlatformStorageSpecBackendGCS) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpecBackendGCS{}
	}

	var errs []error

	if s.GetProjectID() == "" {
		errs = append(errs, shared.PrefixResourceErrors("projectID", errors.New("must be not empty")))
	}

	errs = append(errs,
		shared.PrefixResourceErrors("credentialsSecret", s.GetCredentialsSecret().Validate()),
		shared.PrefixResourceError("bucketName", shared.ValidateRequired(s.BucketName, shared.ValidateResourceName)),
	)

	return shared.WithErrors(errs...)
}

func (s *ArangoPlatformStorageSpecBackendGCS) GetProjectID() string {
	if s == nil || s.ProjectID == nil {
		return ""
	}
	return *s.ProjectID
}

func (s *ArangoPlatformStorageSpecBackendGCS) GetBucketName() string {
	if s == nil || s.BucketName == nil {
		return ""
	}
	return *s.BucketName
}

func (s *ArangoPlatformStorageSpecBackendGCS) GetBucketPrefix() string {
	if s == nil || s.BucketPrefix == nil {
		return ""
	}
	return *s.BucketPrefix
}

func (s *ArangoPlatformStorageSpecBackendGCS) GetCredentialsSecret() *sharedApi.Object {
	if s == nil || s.CredentialsSecret == nil {
		return &sharedApi.Object{}
	}
	return s.CredentialsSecret
}
