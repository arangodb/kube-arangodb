//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	goStrings "strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
	utilTokenLoader "github.com/arangodb/kube-arangodb/pkg/util/token/loader"
)

func NewJWTAuthentication(path string) Authenticator {
	return &jwtAuthentication{
		cache: cache.NewObject(utilTokenLoader.SecretCacheDirectory(path, 10*time.Second)),
	}
}

type jwtAuthentication struct {
	cache cache.Object[utilToken.Secret]
}

func (j *jwtAuthentication) ValidateGRPC(ctx context.Context) (*Identity, error) {
	secret, err := j.cache.Get(ctx)
	if err != nil {
		return nil, err
	}

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
		z := goStrings.SplitN(h, " ", 2)
		if len(z) != 2 {
			return nil, nil
		}

		if goStrings.ToLower(z[0]) != "bearer" {
			return nil, nil
		}

		user, roles, _, err := secret.Details(z[1])
		if err != nil {
			return nil, err
		}

		return &Identity{
			User:  user,
			Roles: roles,
		}, nil

	default:
		return nil, nil
	}
}
