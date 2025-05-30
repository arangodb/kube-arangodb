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
	"fmt"
	goHttp "net/http"
	"time"
)

type MaintenanceClient interface {
	EnableMaintenance(ctx context.Context, id string, timeout Seconds) error
	EnableMaintenanceWithDefaultTimeout(ctx context.Context, id string) error
	DisableMaintenance(ctx context.Context, id string) error
}

type MemberMaintenanceMode string

const (
	MemberMaintenanceModeMaintenance MemberMaintenanceMode = "maintenance"
	MemberMaintenanceModeNormal      MemberMaintenanceMode = "normal"

	MemberMaintenanceUrl = "/_admin/cluster/maintenance/%s"

	DefaultMaintenanceModeTimeout = Seconds(15 * time.Minute)
)

type MemberMaintenanceRequest struct {
	Mode    MemberMaintenanceMode `json:"mode"`
	Timeout *Seconds              `json:"timeout,omitempty"`
}

func (c *client) EnableMaintenance(ctx context.Context, id string, timeout Seconds) error {
	return c.setMaintenance(ctx, id, MemberMaintenanceModeMaintenance, timeout.Ptr())
}

func (c *client) EnableMaintenanceWithDefaultTimeout(ctx context.Context, id string) error {
	return c.EnableMaintenance(ctx, id, DefaultMaintenanceModeTimeout)
}

func (c *client) DisableMaintenance(ctx context.Context, id string) error {
	return c.setMaintenance(ctx, id, MemberMaintenanceModeNormal, nil)
}

func (c *client) setMaintenance(ctx context.Context, id string, mode MemberMaintenanceMode, timeout *Seconds) error {
	req, err := c.c.NewRequest(goHttp.MethodPut, fmt.Sprintf(MemberMaintenanceUrl, id))
	if err != nil {
		return err
	}

	if r, err := req.SetBody(MemberMaintenanceRequest{
		Mode:    mode,
		Timeout: timeout,
	}); err != nil {
		return err
	} else {
		req = r
	}

	resp, err := c.c.Do(ctx, req)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(goHttp.StatusOK); err != nil {
		return err
	}

	return nil
}
