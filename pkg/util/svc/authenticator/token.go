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

	"google.golang.org/grpc/metadata"

	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func NewTokenAuthenticator(token string) Authenticator {
	return &tokenAuthenticator{token: token}
}

type tokenAuthenticator struct {
	token string
}

func (t tokenAuthenticator) ValidateGRPC(ctx context.Context) (*Identity, error) {
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

		if strings.ToLower(z[0]) != "token" {
			return nil, nil
		}

		if z[1] != t.token {
			return nil, nil
		}

		return &Identity{}, nil

	default:
		return nil, nil
	}
}
