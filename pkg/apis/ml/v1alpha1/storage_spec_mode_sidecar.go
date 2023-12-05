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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLStorageSpecModeSidecar struct {
	// ListenPort defines on which port the sidecar container will be listening for connections
	// +doc/default: 9201
	ListenPort *uint16 `json:"listenPort,omitempty"`

	// Image define default image used for the extension
	*sharedApi.Image `json:",inline"`

	// Resources holds resource requests & limits for container running the S3 proxy
	*sharedApi.Resources `json:",inline"`
}

func (s *ArangoMLStorageSpecModeSidecar) GetImage() *sharedApi.Image {
	if s == nil || s.Image == nil {
		return nil
	}

	return s.Image
}

func (s *ArangoMLStorageSpecModeSidecar) GetResources() *sharedApi.Resources {
	if s == nil || s.Resources == nil {
		return nil
	}

	return s.Resources
}

func (s *ArangoMLStorageSpecModeSidecar) Validate() error {
	if s == nil {
		s = &ArangoMLStorageSpecModeSidecar{}
	}

	var err []error

	if s.GetListenPort() < 1 {
		err = append(err, shared.PrefixResourceErrors("database", errors.Newf("must be positive")))
	}

	err = append(err, s.GetResources().Validate())

	return shared.WithErrors(err...)
}

func (s *ArangoMLStorageSpecModeSidecar) GetListenPort() uint16 {
	if s == nil || s.ListenPort == nil {
		return 9201
	}
	return *s.ListenPort
}
