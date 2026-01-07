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
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func handler(mods ...util.ModR[Configuration]) *implementation {
	return newInternal(NewConfiguration().With(mods...))
}

func Server(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) svc.ServiceStarter {
	auth := authenticator.NewBasicAuthenticator(cache.NewObject(func(ctx context.Context) (map[string]string, time.Duration, error) {
		return map[string]string{
			"root": "test",
		}, time.Hour, nil
	}))

	var m []util.ModR[Configuration]
	m = append(m, func(in Configuration) Configuration {
		in.Authenticator = auth
		return in
	})
	m = append(m, mods...)

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
			MuxExtensions: []runtime.ServeMuxOption{
				runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   false,
						EmitUnpopulated: false,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				}),
			},
		},
		Wrap: svc.RequestWraps{
			metrics.Wrapper,
		},
	}, handler(m...))
	require.NoError(t, err)

	return local.Start(ctx)
}

func AuthenticatedContext(t *testing.T, username, password string) context.Context {
	return metadata.NewOutgoingContext(t.Context(), metadata.Pairs(
		"authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))),
	)
}
