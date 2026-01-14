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

package license_manager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	goHttp "net/http"
	"path"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arangodb/kube-arangodb/pkg/platform/inventory"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

const (
	ArangoLicenseManagerEndpoint = "license.arango.ai"
)

func NewClient(endpoint, id, key string, mods ...util.Mod[goHttp.Transport]) Client {
	c := operatorHTTP.NewHTTPClient(
		operatorHTTP.WithTransport(mods...),
		operatorHTTP.DoNotFollowRedirects,
	)

	return client{
		endpoint: fmt.Sprintf("https://%s", endpoint),
		client:   c,
		creds: clientCredentials{
			username: id,
			password: key,
		},
	}
}

type client struct {
	endpoint string

	client operatorHTTP.HTTPClient

	creds clientCredentials
}

func (c client) url(parts ...string) string {
	if len(parts) == 0 {
		return c.endpoint
	}
	return path.Join(c.endpoint, path.Join(parts...))
}

type clientCredentials struct {
	username, password string
}

func (c clientCredentials) Authenticate(in *goHttp.Request) {
	in.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password)))
}

type Client interface {
	Identity(ctx context.Context) (Identity, error)

	License(ctx context.Context, req LicenseRequest) (LicenseResponse, error)

	Registry(ctx context.Context) (RegistryResponse, error)
	RegistryConfig(ctx context.Context, endpoint, id string, token *string, stages ...Stage) ([]byte, error)
}

type LicenseRequest struct {
	DeploymentID *string                             `json:"deployment_id,omitempty"`
	TTL          *ugrpc.Object[*durationpb.Duration] `json:"ttl,omitempty"`
	Inventory    *ugrpc.Object[*inventory.Spec]      `json:"inventory,omitempty"`
}

type LicenseResponse struct {
	ID      string                                `json:"id"`
	License string                                `json:"license"`
	Expires *ugrpc.Object[*timestamppb.Timestamp] `json:"expires,omitempty"`
}

type Identity struct {
}

type RegistryResponse struct {
	Token string `json:"token"`
}

func (c client) Identity(ctx context.Context) (Identity, error) {
	return operatorHTTP.Get[Identity, *ugrpc.RequestError](ctx, c.client, c.url("_api", "v1", "identity"), c.creds.Authenticate).WithCode(200).Get()
}

func (c client) RegistryConfig(ctx context.Context, endpoint, id string, token *string, stages ...Stage) ([]byte, error) {
	var t string

	if token != nil {
		t = *token
	} else {
		tk, err := c.Registry(ctx)
		if err != nil {
			return nil, err
		}
		t = tk.Token
	}

	r, err := NewRegistryAuth(endpoint, id, t, stages...)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c client) License(ctx context.Context, req LicenseRequest) (LicenseResponse, error) {
	return operatorHTTP.Post[LicenseRequest, LicenseResponse, *ugrpc.RequestError](ctx, c.client, req, c.url("_api", "v1", "license"), c.creds.Authenticate).WithCode(200).Get()
}

func (c client) Registry(ctx context.Context) (RegistryResponse, error) {
	return operatorHTTP.Get[RegistryResponse, *ugrpc.RequestError](ctx, c.client, c.url("_api", "v1", "token"), c.creds.Authenticate).WithCode(200).Get()
}
