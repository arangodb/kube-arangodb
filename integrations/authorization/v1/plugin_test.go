//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
)

func newPluginTest() pluginTest {
	return &pluginTestImpl{
		responses: map[string]*pbAuthorizationV1.AuthorizationV1PermissionResponse{},
	}
}

type pluginTest interface {
	pbImplAuthorizationV1Shared.Plugin

	Set(t *testing.T, req *pbAuthorizationV1.AuthorizationV1PermissionRequest, resp *pbAuthorizationV1.AuthorizationV1PermissionResponse)
}

type pluginTestImpl struct {
	lock sync.Mutex

	responses map[string]*pbAuthorizationV1.AuthorizationV1PermissionResponse
}

func (p *pluginTestImpl) Revision() uint64 {
	return 0
}

func (p *pluginTestImpl) Background(ctx context.Context) {

}

func (p *pluginTestImpl) Ready(ctx context.Context) error {
	return nil
}

func (p *pluginTestImpl) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if r, ok := p.responses[req.Hash()]; ok {
		return r, nil
	}

	return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Default Response",
		Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
	}, nil
}

func (p *pluginTestImpl) Set(t *testing.T, req *pbAuthorizationV1.AuthorizationV1PermissionRequest, resp *pbAuthorizationV1.AuthorizationV1PermissionResponse) {
	p.lock.Lock()
	defer p.lock.Unlock()

	require.NotNil(t, req)

	if resp == nil {
		delete(p.responses, req.Hash())
	} else {
		p.responses[req.Hash()] = resp
	}
}
