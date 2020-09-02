//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package agency

import (
	"context"
	"net/http"

	"github.com/arangodb/go-driver"
)

type Maintenance struct {
	Result string `json:"result"`
}

func (m Maintenance) Enabled() bool {
	return m.Result == "Maintenance"
}

func GetMaintenanceMode(ctx context.Context, client driver.Client) (Maintenance, error) {
	conn := client.Connection()
	r, err := conn.NewRequest(http.MethodGet, "/_admin/cluster/maintenance")
	if err != nil {
		return Maintenance{}, err
	}

	resp, err := conn.Do(ctx, r)
	if err != nil {
		return Maintenance{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return Maintenance{}, err
	}

	var m Maintenance

	if err := resp.ParseBody("", &m); err != nil {
		return Maintenance{}, err
	}

	return m, nil
}

func SetMaintenanceMode(ctx context.Context, client driver.Client, enabled bool) error {
	data := "on"
	if !enabled {
		data = "off"
	}

	conn := client.Connection()
	r, err := conn.NewRequest(http.MethodPut, "/_admin/cluster/maintenance")
	if err != nil {
		return err
	}

	if _, err := r.SetBody(data); err != nil {
		return err
	}

	resp, err := conn.Do(ctx, r)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return err
	}

	return nil
}
