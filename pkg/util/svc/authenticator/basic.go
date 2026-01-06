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

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func NewBasicAuthenticator(object cache.Object[map[string]string]) Authenticator {
	return &basicAuthenticator{object: object}
}

type basicAuthenticator struct {
	object cache.Object[map[string]string]
}

func (b *basicAuthenticator) Init(ctx context.Context) error {
	_, err := b.object.Get(ctx)
	return err
}

func (b *basicAuthenticator) ValidateGRPC(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	v := md.Get("authorization")
	switch len(v) {
	case 0:
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")

	case 1:
		h := v[0]
		if goStrings.HasPrefix(h, "Basic ") {
			h = goStrings.TrimPrefix(h, "Basic ")
			c, err := base64.StdEncoding.DecodeString(h)
			if err == nil {
				if n := goStrings.SplitN(string(c), ":", 2); len(n) == 2 {
					if b.validate(ctx, n[0], n[1]) == nil {
						break
					}
				}
			}
		}
		return status.Errorf(codes.Unauthenticated, "authorization token is invalid")

	default:
		return status.Errorf(codes.Unauthenticated, "authorization token is invalid")
	}

	return nil
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
