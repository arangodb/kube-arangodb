//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package impl

import (
	"context"
	"encoding/base64"
	"fmt"
	goHttp "net/http"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/api/server"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Version(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	client := operatorHTTP.NewHTTPClient()

	require.NoError(t, ugrpc.Get[*server.Version](ctx, client, fmt.Sprintf("http://%s/_api/version", q.HTTPAddress())).WithCode(goHttp.StatusUnauthorized).Validate())
	require.NoError(t, ugrpc.Get[*server.Version](ctx, client, fmt.Sprintf("http://%s/_api/version", q.HTTPAddress()), func(in *goHttp.Request) {
		in.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "root2", "test")))))
	}).WithCode(goHttp.StatusUnauthorized).Validate())
	require.NoError(t, ugrpc.Get[*server.Version](ctx, client, fmt.Sprintf("http://%s/_api/version", q.HTTPAddress()), func(in *goHttp.Request) {
		in.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "root", "test")))))
	}).WithCode(goHttp.StatusOK).Validate())

	gclient := tgrpc.NewGRPCClient(t, ctx, server.NewOperatorClient, q.Address())

	_, err := gclient.GetVersion(t.Context(), &pbSharedV1.Empty{})
	tgrpc.AsGRPCError(t, err).Code(t, codes.Unauthenticated)

	_, err = gclient.GetVersion(AuthenticatedContext(t, "root2", "test"), &pbSharedV1.Empty{})
	tgrpc.AsGRPCError(t, err).Code(t, codes.Unauthenticated)

	_, err = gclient.GetVersion(AuthenticatedContext(t, "root", "test"), &pbSharedV1.Empty{})
	require.NoError(t, err)
}

func Test_Metrics(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	resp, err := goHttp.Get(fmt.Sprintf("http://%s/metrics", q.HTTPAddress()))
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)

	defer resp.Body.Close()

	mfChan := make(chan *dto.MetricFamily, 2048)

	go func() {
		if err := prom2json.ParseReader(resp.Body, mfChan); err != nil {
			require.NoError(t, err)
		}
	}()

	for mf := range mfChan {
		prom2json.NewFamily(mf)
	}
}

func Test_Health(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	resp, err := goHttp.Get(fmt.Sprintf("http://%s/health", q.HTTPAddress()))
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)
}

func Test_Ready(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	resp, err := goHttp.Get(fmt.Sprintf("http://%s/ready", q.HTTPAddress()))
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)
}
