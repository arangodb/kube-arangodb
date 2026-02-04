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
	"fmt"
	goHttp "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
)

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()

	client, _ := Client(t, ctx, Handler(p))

	resp, err := client.Evaluate(ctx, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Action:   "test:Get",
		Resource: "test",
	})
	require.NoError(t, err)
	require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetEffect())
}

func Test_ServiceHTTP(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()

	_, endpoint := Client(t, ctx, Handler(p))

	data, err := http.Post[ugrpc.Object[*pbAuthorizationV1.AuthorizationV1PermissionRequest], any, error](ctx, goHttp.DefaultClient, ugrpc.NewObject(&pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Action:   "test:Get",
		Resource: "test",
	}), fmt.Sprintf("http://%s/_integration/authorization/v1/evaluate", endpoint)).WithCode(200).Data()
	require.NoError(t, err)

	t.Logf("Response: %s", string(data))
}
