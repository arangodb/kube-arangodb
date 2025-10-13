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

package manager

import (
	"context"
	"fmt"
	goHttp "net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func NewClient(endpoint, id, key string, mods ...util.Mod[goHttp.Transport]) (Client, error) {
	transport := operatorHTTP.Transport(mods...)

	stageEndpoint := fmt.Sprintf("https://%s", endpoint)

	connConfig := http.ConnectionConfig{
		Transport:          transport,
		DontFollowRedirect: true,
		Endpoints:          []string{stageEndpoint},
	}

	conn, err := http.NewConnection(connConfig)
	if err != nil {
		return nil, err
	}

	conn, err = conn.SetAuthentication(driver.BasicAuthentication(id, key))
	if err != nil {
		return nil, err
	}

	return NewClientFromConn(conn), nil
}

func NewClientFromConn(conn driver.Connection) Client {
	return client{
		conn: conn,
	}
}

type Client interface {
	License(ctx context.Context, req LicenseRequest) (LicenseResponse, error)

	Registry(ctx context.Context) (RegistryResponse, error)
}

type LicenseRequest struct {
	DeploymentID *string        `json:"deployment_id,omitempty"`
	TTL          *time.Duration `json:"ttl,omitempty"`
	Inventory    *Inventory     `json:"inventory,omitempty"`
}

type LicenseResponse struct {
	ID      string `json:"id"`
	License string `json:"license"`
}

type RegistryResponse struct {
	Token string `json:"token"`
}

type client struct {
	conn driver.Connection
}

func (c client) License(ctx context.Context, req LicenseRequest) (LicenseResponse, error) {
	return arangod.PostRequest[LicenseRequest, LicenseResponse](ctx, c.conn, req, "_api", "v1", "license").AcceptCode(200).Response()
}

func (c client) Registry(ctx context.Context) (RegistryResponse, error) {
	return arangod.GetRequest[RegistryResponse](ctx, c.conn, "_api", "v1", "registry", "token").AcceptCode(200).Response()
}
