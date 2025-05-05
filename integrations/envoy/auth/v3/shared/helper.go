//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type Token string

type Helper[K comparable] cache.Cache[K, Token]

type HelperInterface[K comparable] interface {
	Token(ctx context.Context, in K) (Token, error)
}

type HelperFunc[K comparable] func(ctx context.Context, in K) (Token, error)

func NewHelperInterface[K comparable](f HelperInterface[K]) Helper[K] {
	return NewHelper[K](f.Token)
}

func NewHelper[K comparable](f HelperFunc[K]) Helper[K] {
	return cache.NewCache(cache.CacheExtract[K, Token](f), DefaultTTL)
}
