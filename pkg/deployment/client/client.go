//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package client

import (
	"context"
	goHttp "net/http"
	"time"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

func NewClient(c adbDriverV2Connection.Connection) Client {
	return &client{
		c: c,
	}
}

type Client interface {
	LicenseClient
	MaintenanceClient
	RebalanceClient

	GetTLS(ctx context.Context) (TLSDetails, error)
	RefreshTLS(ctx context.Context) (TLSDetails, error)

	GetEncryption(ctx context.Context) (EncryptionDetails, error)
	RefreshEncryption(ctx context.Context) (EncryptionDetails, error)

	GetJWT(ctx context.Context) (JWTDetails, error)
	RefreshJWT(ctx context.Context) (JWTDetails, error)

	DeleteExpiredJobs(ctx context.Context, timeout time.Duration) error

	Compact(ctx context.Context, request *CompactRequest) error

	Inventory(ctx context.Context) (Inventory, error)

	DeploymentID(ctx context.Context) (DeploymentID, error)
}

type client struct {
	c adbDriverV2Connection.Connection
}

func (c *client) GetTLS(ctx context.Context) (TLSDetails, error) {
	return arangod.GetRequest[TLSDetails](ctx, c.c, "_admin", "server", "tls").AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) RefreshTLS(ctx context.Context) (TLSDetails, error) {
	return arangod.PostRequest[any, TLSDetails](ctx, c.c, nil, "_admin", "server", "tls").AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) GetEncryption(ctx context.Context) (EncryptionDetails, error) {
	return arangod.GetRequest[EncryptionDetails](ctx, c.c, "_admin", "server", "encryption").AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) RefreshEncryption(ctx context.Context) (EncryptionDetails, error) {
	return arangod.PostRequest[any, EncryptionDetails](ctx, c.c, nil, "_admin", "server", "encryption").AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) GetJWT(ctx context.Context) (JWTDetails, error) {
	return arangod.GetRequest[JWTDetails](ctx, c.c, "_admin", "server", "jwt").AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) RefreshJWT(ctx context.Context) (JWTDetails, error) {
	return arangod.PostRequest[any, JWTDetails](ctx, c.c, nil, "_admin", "server", "jwt").AcceptCode(goHttp.StatusOK).Response()
}
