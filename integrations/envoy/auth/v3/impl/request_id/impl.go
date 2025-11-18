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

package request_id

import (
	"context"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	return impl{}, true
}

type impl struct {
}

func (a impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	var header = pbEnvoyCoreV3.HeaderValueOption{
		Header: &pbEnvoyCoreV3.HeaderValue{
			Key:   utilConstants.EnvoyRequestIDHeader,
			Value: string(uuid.NewUUID()),
		},
	}
	current.Headers = append(current.Headers, &header)
	current.ResponseHeaders = append(current.ResponseHeaders, &header)

	return nil
}
