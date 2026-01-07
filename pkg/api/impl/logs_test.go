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

package impl

import (
	"context"
	"encoding/base64"
	"fmt"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	goHttp "net/http"
	"testing"
)

const loggerName = "testing-internal"

var logger = logging.Global().RegisterAndGetLogger(loggerName, logging.Info)

func Test_Logs(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	client := operatorHTTP.NewHTTPClient()

	require.NoError(t, ugrpc.Get[*server.LogLevelConfig](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress())).WithCode(goHttp.StatusUnauthorized).Validate())
	require.NoError(t, ugrpc.Get[*server.LogLevelConfig](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), func(in *goHttp.Request) {
		in.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "root2", "test")))))
	}).WithCode(goHttp.StatusUnauthorized).Validate())
	require.NoError(t, ugrpc.Get[*server.LogLevelConfig](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), func(in *goHttp.Request) {
		in.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "root", "test")))))
	}).WithCode(goHttp.StatusOK).Validate())

	gclient := tgrpc.NewGRPCClient(t, ctx, server.NewOperatorClient, q.Address())

	_, err := gclient.GetLogLevel(t.Context(), &pbSharedV1.Empty{})
	tgrpc.AsGRPCError(t, err).Code(t, codes.Unauthenticated)

	_, err = gclient.GetLogLevel(AuthenticatedContext(t, "root2", "test"), &pbSharedV1.Empty{})
	tgrpc.AsGRPCError(t, err).Code(t, codes.Unauthenticated)

	_, err = gclient.GetLogLevel(AuthenticatedContext(t, "root", "test"), &pbSharedV1.Empty{})
	require.NoError(t, err)
}

func Test_Logs_Change(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Server(t, ctx)

	client := operatorHTTP.NewHTTPClient()

	type data struct {
		Topics map[string]string `json:"topics"`
	}

	auth := func(in *goHttp.Request) {
		in.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "root", "test")))))
	}

	res, err := ugrpc.AsJSON[*pbSharedV1.Empty, data](ugrpc.Get[*pbSharedV1.Empty](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), auth).WithCode(goHttp.StatusOK))
	require.NoError(t, err)

	require.Contains(t, res.Topics, loggerName)

	ugrpc.Post[*server.LogLevelConfig, *pbSharedV1.Empty](ctx, client, &server.LogLevelConfig{
		Topics: map[string]server.LogLevel{
			loggerName: server.LogLevel_fatal,
		},
	}, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), auth).WithCode(goHttp.StatusOK)

	res, err = ugrpc.AsJSON[*pbSharedV1.Empty, data](ugrpc.Get[*pbSharedV1.Empty](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), auth).WithCode(goHttp.StatusOK))
	require.NoError(t, err)
	require.Contains(t, res.Topics, loggerName)
	require.Equal(t, res.Topics[loggerName], server.LogLevel_fatal.String())

	ugrpc.Post[*server.LogLevelConfig, *pbSharedV1.Empty](ctx, client, &server.LogLevelConfig{
		Topics: map[string]server.LogLevel{
			loggerName: server.LogLevel_info,
		},
	}, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), auth).WithCode(goHttp.StatusOK)

	res, err = ugrpc.AsJSON[*pbSharedV1.Empty, data](ugrpc.Get[*pbSharedV1.Empty](ctx, client, fmt.Sprintf("http://%s/log/level", q.HTTPAddress()), auth).WithCode(goHttp.StatusOK))
	require.NoError(t, err)
	require.Contains(t, res.Topics, loggerName)
	require.Equal(t, res.Topics[loggerName], server.LogLevel_info.String())
}
