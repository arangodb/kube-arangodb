//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"net/http"
)

const AdminLicenseUrl = "/_admin/license"

type LicenseClient interface {
	GetLicense(ctx context.Context) (License, error)
	SetLicense(ctx context.Context, license string, force bool) error
}

type License struct {
	Hash string `json:"hash,omitempty"`
}

func (c *client) GetLicense(ctx context.Context) (License, error) {
	req, err := c.c.NewRequest(http.MethodGet, AdminLicenseUrl)
	if err != nil {
		return License{}, err
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return License{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return License{}, err
	}

	var l License

	if err := resp.ParseBody("", &l); err != nil {
		return License{}, err
	}

	return l, nil
}

func (c *client) SetLicense(ctx context.Context, license string, force bool) error {
	req, err := c.c.NewRequest(http.MethodPut, AdminLicenseUrl)
	if err != nil {
		return err
	}

	if r, err := req.SetBody(license); err != nil {
		return err
	} else {
		req = r
	}

	if force {
		req = req.SetQuery("force", "true")
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(http.StatusCreated); err != nil {
		return err
	}

	return nil
}
