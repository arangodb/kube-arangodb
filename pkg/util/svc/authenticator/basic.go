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

package authenticator

import (
	"context"
	"encoding/base64"
	goStrings "strings"

	"google.golang.org/grpc/metadata"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func NewBasicAuthenticator(object cache.Object[map[string]string]) Authenticator {
	return &basicAuthenticator{object: object}
}

type basicAuthenticator struct {
	object cache.Object[map[string]string]
}

func (b *basicAuthenticator) ValidateGRPC(ctx context.Context) (*Identity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil
	}

	v := md.Get("authorization")
	switch len(v) {
	case 0:
		return nil, nil

	case 1:
		h := v[0]
		z := strings.SplitN(h, " ", 2)
		if len(z) != 2 {
			return nil, nil
		}

		if strings.ToLower(z[0]) != "basic" {
			return nil, nil
		}

		c, err := base64.StdEncoding.DecodeString(z[1])
		if err == nil {
			if n := goStrings.SplitN(string(c), ":", 2); len(n) == 2 {
				if b.validate(ctx, n[0], n[1]) == nil {
					return &Identity{User: util.NewType(n[0])}, nil
				}
			}
		}

		return nil, nil

	default:
		return nil, nil
	}
}

func (b *basicAuthenticator) validate(ctx context.Context, username, password string) error {
	data, err := b.object.Get(ctx)
	if err != nil {
		return err
	}

	if v, ok := data[username]; !ok || v != password {
		return errors.Errorf("username or password is incorrect")
	}

	return nil
}
