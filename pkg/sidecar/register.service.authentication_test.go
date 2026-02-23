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

package sidecar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	sidecarSvcAuthnDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authentication/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func Test_ServiceClientAuthentication(t *testing.T) {
	tm := tests.NewTokenManager(t)

	token1 := tests.GenerateECDSAP256Token(t)
	token2 := tests.GenerateECDSAP256Token(t)

	tm.Set(t, token1, token2)

	t.Run("No Authentication", func(t *testing.T) {
		runSidecar(t).
			run("Check if can get data", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(c)

				tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{}).Code(t, codes.Unavailable)
			})
	})

	t.Run("With Authentication", func(t *testing.T) {
		runSidecar(t).
			addArgs("--sidecar.auth", tm.Path()).
			run("Check if can get data without auth", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(c)

				tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{}).Code(t, codes.Unauthenticated)
			}).
			run("Check if can get data with auth", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(c)

				resp := tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{}).Get(t)
				require.Len(t, resp.Keys, 2)
			}).
			run("Check if can get data with auth and one token", func(t *testing.T) {
				tm.Set(t, token1)

				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(c)

				resp := tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{}).Get(t)
				require.Len(t, resp.Keys, 1)
			}).
			run("Check if can get data with auth and refresh", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(c)

				resp := tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{}).Get(t)
				require.Len(t, resp.Keys, 1)

				tgrpc.NewExecutor(t, client.GetOptionalKeys, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{
					Checksum: util.NewType(resp.GetChecksum()),
				}).Code(t, codes.AlreadyExists)
			})
	})
}
