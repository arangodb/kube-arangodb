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

package v1alpha1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoPlatformStorageSpec struct {
	// Deployment specifies the ArangoDeployment object name
	Deployment *string `json:"deployment,omitempty"`

	// BucketName specifies the name of the bucket
	// Required
	BucketName *string `json:"bucketName,omitempty"`

	// BucketPath specifies the path within the bucket
	// +doc/default: /
	BucketPath *string `json:"bucketPath,omitempty"`

	// Mode defines how storage implementation should be deployed
	Mode *ArangoPlatformStorageSpecMode `json:"mode,omitempty"`
	// Backend defines how storage is implemented
	Backend *ArangoPlatformStorageSpecBackend `json:"backend,omitempty"`
}

func (s *ArangoPlatformStorageSpec) GetDeployment() string {
	if s == nil || s.Deployment == nil {
		return ""
	}

	return *s.Deployment
}

func (s *ArangoPlatformStorageSpec) GetBucketName() string {
	if s == nil || s.BucketName == nil {
		return ""
	}
	return *s.BucketName
}

func (s *ArangoPlatformStorageSpec) GetBucketPath() string {
	if s == nil || s.BucketPath == nil {
		return "/"
	}
	return *s.BucketPath
}

func (s *ArangoPlatformStorageSpec) GetMode() *ArangoPlatformStorageSpecMode {
	if s == nil || s.Mode == nil {
		return nil
	}
	return s.Mode
}

func (s *ArangoPlatformStorageSpec) GetBackend() *ArangoPlatformStorageSpecBackend {
	if s == nil || s.Backend == nil {
		return nil
	}
	return s.Backend
}

func (s *ArangoPlatformStorageSpec) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpec{}
	}

	if err := shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceError("backend", s.Backend.Validate()),
		shared.PrefixResourceError("mode", s.Mode.Validate()),
		shared.PrefixResourceError("bucket", shared.ValidateRequired(s.BucketName, shared.ValidateResourceName)),
	)); err != nil {
		return err
	}

	return nil
}
