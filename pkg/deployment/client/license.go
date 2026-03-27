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

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

type LicenseClient interface {
	GetLicense(ctx context.Context) (License, error)
	SetLicense(ctx context.Context, license string, force bool) error
}

type License struct {
	Hash string `json:"hash,omitempty"`

	Features *LicenseFeatures `json:"features,omitempty"`
}

type LicenseFeatures struct {
	Expires *int64 `json:"expires,omitempty"`
}

func (l *License) Expires() time.Time {
	if l == nil || l.Features == nil || l.Features.Expires == nil {
		return time.Time{}
	}
	return time.Unix(*l.Features.Expires, 0).UTC()
}

func (c *client) GetLicense(ctx context.Context) (License, error) {
	return arangod.GetRequest[License](ctx, c.c, "_admin", "license").Do(ctx).AcceptCode(goHttp.StatusOK).Response()
}

func (c *client) SetLicense(ctx context.Context, license string, force bool) error {
	req := arangod.PutRequest[string, License](ctx, c.c, license, "_admin", "license")

	if force {
		req = req.Query("force", "true")
	}
	return req.
		Do(ctx).
		AcceptCode(goHttp.StatusCreated).
		Evaluate()
}
