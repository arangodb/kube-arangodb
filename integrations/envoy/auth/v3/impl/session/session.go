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

package session

import (
	"context"
	"fmt"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver/v2/arangodb"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func NewManager[T any](ctx context.Context, t Type, client cache.Object[arangodb.Collection]) Manager[T] {
	return manager[T]{
		t:     t,
		cache: cache.NewRemoteCache[*session](client),
	}
}

type Manager[T any] interface {
	Put(ctx context.Context, expires time.Time, obj T) (string, error)

	Get(ctx context.Context, key string) (T, bool, time.Duration, error)

	Invalidate(ctx context.Context, key string) (bool, error)
}

type manager[T any] struct {
	t     Type
	cache cache.RemoteCache[*session]
}

func (m manager[T]) Put(ctx context.Context, expires time.Time, obj T) (string, error) {
	data, err := sharedApi.NewAny(obj)
	if err != nil {
		return "", err
	}

	key := string(uuid.NewUUID())
	hkey := util.SHA256FromString(key)

	return key, m.cache.Put(ctx, fmt.Sprintf("%s_%s", m.t, hkey), &session{
		Key:       fmt.Sprintf("%s_%s", m.t, hkey),
		Object:    data,
		ExpiresAt: meta.NewTime(expires),
	})
}

func (m manager[T]) Get(ctx context.Context, key string) (T, bool, time.Duration, error) {
	ret, ok, err := m.cache.Get(ctx, fmt.Sprintf("%s_%s", m.t, util.SHA256FromString(key)))
	if err != nil {
		return util.Default[T](), false, 0, err
	}

	if !ok {
		return util.Default[T](), false, 0, nil
	}

	obj, err := sharedApi.FromAny[T](ret.Object)
	if err != nil {
		return util.Default[T](), false, 0, nil
	}

	return obj, true, time.Until(ret.Expires()), nil
}

func (m manager[T]) Invalidate(ctx context.Context, key string) (bool, error) {
	return m.cache.Revoke(ctx, fmt.Sprintf("%s_%s", m.t, util.SHA256FromString(key)))
}

type session struct {
	Key string `json:"_key"`

	Object sharedApi.Any `json:"object,omitempty"`

	ExpiresAt meta.Time `json:"expires_at"`
}

func (s *session) SetKey(k string) {
	s.Key = k
}

func (s *session) GetKey() string {
	if s == nil {
		return ""
	}

	return s.Key
}

func (s *session) Expires() time.Time {
	if s == nil {
		return time.Time{}
	}

	return s.ExpiresAt.Time
}
