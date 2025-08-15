//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/go-driver"

	pbInventoryV1 "github.com/arangodb/kube-arangodb/integrations/inventory/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func NewClient(c driver.Connection, log logging.Logger) Client {
	if log != nil {
		c = loggerConnection(c, log)
	}

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

	Inventory(ctx context.Context) (*pbInventoryV1.Inventory, error)
}

type client struct {
	c driver.Connection
}

func (c *client) parseTLSResponse(response driver.Response) (TLSDetails, error) {
	if err := response.CheckStatus(goHttp.StatusOK); err != nil {
		return TLSDetails{}, err
	}

	var d TLSDetails

	if err := response.ParseBody("", &d); err != nil {
		return TLSDetails{}, err
	}

	return d, nil
}

func (c *client) parseEncryptionResponse(response driver.Response) (EncryptionDetails, error) {
	if err := response.CheckStatus(goHttp.StatusOK); err != nil {
		return EncryptionDetails{}, err
	}

	var d EncryptionDetails

	if err := response.ParseBody("", &d); err != nil {
		return EncryptionDetails{}, err
	}

	return d, nil
}

func (c *client) parseJWTResponse(response driver.Response) (JWTDetails, error) {
	if err := response.CheckStatus(goHttp.StatusOK); err != nil {
		return JWTDetails{}, err
	}

	var d JWTDetails

	if err := response.ParseBody("", &d); err != nil {
		return JWTDetails{}, err
	}

	return d, nil
}

func (c *client) GetTLS(ctx context.Context) (TLSDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodGet, "/_admin/server/tls")
	if err != nil {
		return TLSDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return TLSDetails{}, err
	}

	d, err := c.parseTLSResponse(response)
	if err != nil {
		return TLSDetails{}, err
	}

	return d, nil
}

func (c *client) RefreshTLS(ctx context.Context) (TLSDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodPost, "/_admin/server/tls")
	if err != nil {
		return TLSDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return TLSDetails{}, err
	}

	d, err := c.parseTLSResponse(response)
	if err != nil {
		return TLSDetails{}, err
	}

	return d, nil
}

func (c *client) GetEncryption(ctx context.Context) (EncryptionDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodGet, "/_admin/server/encryption")
	if err != nil {
		return EncryptionDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return EncryptionDetails{}, err
	}

	d, err := c.parseEncryptionResponse(response)
	if err != nil {
		return EncryptionDetails{}, err
	}

	return d, nil
}

func (c *client) RefreshEncryption(ctx context.Context) (EncryptionDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodPost, "/_admin/server/encryption")
	if err != nil {
		return EncryptionDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return EncryptionDetails{}, err
	}

	d, err := c.parseEncryptionResponse(response)
	if err != nil {
		return EncryptionDetails{}, err
	}

	return d, nil
}

func (c *client) GetJWT(ctx context.Context) (JWTDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodGet, "/_admin/server/jwt")
	if err != nil {
		return JWTDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return JWTDetails{}, err
	}

	d, err := c.parseJWTResponse(response)
	if err != nil {
		return JWTDetails{}, err
	}

	return d, nil
}

func (c *client) RefreshJWT(ctx context.Context) (JWTDetails, error) {
	r, err := c.c.NewRequest(goHttp.MethodPost, "/_admin/server/jwt")
	if err != nil {
		return JWTDetails{}, err
	}

	response, err := c.c.Do(ctx, r)
	if err != nil {
		return JWTDetails{}, err
	}

	d, err := c.parseJWTResponse(response)
	if err != nil {
		return JWTDetails{}, err
	}

	return d, nil
}
