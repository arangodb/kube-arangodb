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

//go:build testing

package tests

import (
	"context"
	goStrings "strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

const (
	TEST_ARANGODB_ENDPOINT util.EnvironmentVariable = "TEST_ARANGODB_ENDPOINT"
	TEST_ARANGODB_AUTH     util.EnvironmentVariable = "TEST_ARANGODB_AUTH"
)

type ArangoDBTestConfig struct {
	Endpoint string
	Auth     ArangoDBTestConfigAuth
}

func (a ArangoDBTestConfig) Client(t *testing.T) adbDriverV2.Client {
	client := adbDriverV2.NewClient(adbDriverV2Connection.NewHttpConnection(adbDriverV2Connection.HttpConfiguration{
		Authentication: a.Auth.Auth(),
		Endpoint: adbDriverV2Connection.NewRoundRobinEndpoints([]string{
			a.Endpoint,
		}),
		ContentType:    adbDriverV2Connection.ApplicationJSON,
		ArangoDBConfig: adbDriverV2Connection.ArangoDBConfiguration{},
		Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
	}))

	_, err := client.Version(t.Context())
	require.NoError(t, err)

	return client
}

func (a ArangoDBTestConfig) ClientCache() func(ctx context.Context) (adbDriverV2.Client, time.Duration, error) {
	return func(ctx context.Context) (adbDriverV2.Client, time.Duration, error) {
		client := adbDriverV2.NewClient(adbDriverV2Connection.NewHttpConnection(adbDriverV2Connection.HttpConfiguration{
			Authentication: a.Auth.Auth(),
			Endpoint: adbDriverV2Connection.NewRoundRobinEndpoints([]string{
				a.Endpoint,
			}),
			ContentType:    adbDriverV2Connection.ApplicationJSON,
			ArangoDBConfig: adbDriverV2Connection.ArangoDBConfiguration{},
			Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
		}))

		_, err := client.Version(ctx)
		if err != nil {
			return nil, 0, err
		}

		return client, time.Hour, nil
	}
}

type ArangoDBTestConfigAuth struct {
	Basic *ArangoDBTestConfigAuthBasic
	JWT   *ArangoDBTestConfigAuthJWT
}

func (a ArangoDBTestConfigAuth) Auth() adbDriverV2Connection.Authentication {
	if v := a.Basic; v != nil {
		return adbDriverV2Connection.NewBasicAuth(v.Username, v.Password)
	}

	if v := a.JWT; v != nil {
		return adbDriverV2Connection.NewHeaderAuth("authorization", "bearer %s", v.Token)
	}

	return nil
}

type ArangoDBTestConfigAuthBasic struct {
	Username, Password string
}

type ArangoDBTestConfigAuthJWT struct {
	Token string
}

func TestArangoClientCacheFunc(t *testing.T) func(ctx context.Context) (adbDriverV2.Client, time.Duration, error) {
	if !TEST_ARANGODB_ENDPOINT.Exists() {
		return func(ctx context.Context) (adbDriverV2.Client, time.Duration, error) {
			return nil, 0, errors.Errorf("TEST_ARANGODB_ENDPOINT is not set")
		}
	}

	var r ArangoDBTestConfig

	r.Endpoint = TEST_ARANGODB_ENDPOINT.Get()

	r.Auth = TestConfigAuth(t)

	return r.ClientCache()
}

func TestArangoDBConfig(t *testing.T) ArangoDBTestConfig {
	if !TEST_ARANGODB_ENDPOINT.Exists() {
		t.Skipf("TEST_ARANGODB_ENDPOINT is not set")
	}

	var r ArangoDBTestConfig

	r.Endpoint = TEST_ARANGODB_ENDPOINT.Get()

	r.Auth = TestConfigAuth(t)

	return r
}

func TestConfigAuth(t *testing.T) ArangoDBTestConfigAuth {
	auth := TEST_ARANGODB_AUTH.GetOrDefault("none")

	if auth == "none" {
		return ArangoDBTestConfigAuth{}
	}

	parts := goStrings.SplitN(auth, ":", 2)

	require.Len(t, parts, 2)

	switch parts[0] {
	case "basic":
		z := goStrings.SplitN(parts[1], ":", 2)
		require.Len(t, z, 2)

		return ArangoDBTestConfigAuth{
			Basic: &ArangoDBTestConfigAuthBasic{
				Username: z[0],
				Password: z[1],
			},
		}
	case "jwt":
		return ArangoDBTestConfigAuth{JWT: &ArangoDBTestConfigAuthJWT{
			Token: parts[1],
		}}
	}

	require.Fail(t, "Unknown auth type")
	return ArangoDBTestConfigAuth{}
}
