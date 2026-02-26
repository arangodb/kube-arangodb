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

package authentication

import (
	"context"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func NewSecretAuthentication(in cache.Object[utilToken.Secret], mods ...util.ModR[utilToken.Claims]) Authentication {
	return &secretAuthentication{
		in:   in,
		mods: mods,
	}
}

type secretAuthentication struct {
	in cache.Object[utilToken.Secret]

	mods []util.ModR[utilToken.Claims]
}

func (s secretAuthentication) ExtendAuthentication(ctx context.Context) (string, bool, error) {
	secret, err := s.in.Get(ctx)
	if err != nil {
		return "", false, err
	}

	if !secret.Exists() {
		return "", false, nil
	}

	token, err := utilToken.NewClaims().With(s.mods...).Sign(secret)
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("bearer %s", token), true, nil
}
