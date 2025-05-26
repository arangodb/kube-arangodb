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

package cache

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewConfigFile[T any](path string, ttl time.Duration) ConfigFile[T] {
	return &configFile[T]{
		path: path,
		ttl:  ttl,
	}
}

type ConfigFile[T any] interface {
	Get(ctx context.Context) (T, string, error)
}

type configFile[T any] struct {
	lock sync.Mutex

	path string
	hash string
	ttl  time.Duration

	object T

	next time.Time
}

func (c *configFile[T]) Get(_ context.Context) (T, string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if time.Now().After(c.next) {
		d, err := os.ReadFile(c.path)
		if err != nil {
			return util.Default[T](), "", err
		}

		obj, err := util.JsonOrYamlUnmarshal[T](d)
		if err != nil {
			return util.Default[T](), "", err
		}

		c.object = obj
		c.hash = util.SHA256(d)
		c.next = time.Now().Add(c.ttl)
	}

	return c.object, c.hash, nil
}

type HashedConfigurationRetriever[T, S any] func(ctx context.Context, in S) (T, error)

func NewHashedConfiguration[T, S any](config ConfigFile[S], retriever HashedConfigurationRetriever[T, S]) HashedConfiguration[T] {
	return &hashedConfiguration[T, S]{
		config:    config,
		retriever: retriever,
	}
}

type HashedConfiguration[T any] interface {
	Get(ctx context.Context) (T, error)
}

type hashedConfiguration[T, S any] struct {
	lock sync.Mutex

	config ConfigFile[S]

	retriever HashedConfigurationRetriever[T, S]

	hash string
	obj  T
}

func (h *hashedConfiguration[T, S]) Get(ctx context.Context) (T, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	in, hash, err := h.config.Get(ctx)
	if err != nil {
		return util.Default[T](), err
	}

	if h.hash != hash {
		obj, err := h.retriever(ctx, in)
		if err != nil {
			return util.Default[T](), err
		}

		h.hash = hash
		h.obj = obj
	}

	return h.obj, nil
}
