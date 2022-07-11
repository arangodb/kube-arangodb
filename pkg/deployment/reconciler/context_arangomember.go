//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconciler

import (
	"context"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

type ArangoMemberCreateFunc func(obj *api.ArangoMember)
type ArangoMemberUpdateFunc func(obj *api.ArangoMember) bool
type ArangoMemberStatusUpdateFunc func(obj *api.ArangoMember, s *api.ArangoMemberStatus) bool

type ArangoMemberContext interface {
	// WithArangoMember start ArangoMember scope. Used in ACS
	WithArangoMember(cache inspector.Inspector, timeout time.Duration, name string) ArangoMemberModContext

	// WithCurrentArangoMember start ArangoMember scope within current deployment scope
	WithCurrentArangoMember(name string) ArangoMemberModContext
}

func NewArangoMemberModContext(cache inspector.Inspector, timeout time.Duration, name string) ArangoMemberModContext {
	return arangoMemberModContext{
		cache:   cache,
		name:    name,
		timeout: timeout,
	}
}

type ArangoMemberModContext interface {
	// Exists returns true if object exists
	Exists(ctx context.Context) bool
	// Create creates ArangoMember
	Create(ctx context.Context, obj *api.ArangoMember) error
	// Update run action with update of ArangoMember
	Update(ctx context.Context, action ArangoMemberUpdateFunc) error
	// UpdateStatus run action with update of ArangoMember Status
	UpdateStatus(ctx context.Context, action ArangoMemberStatusUpdateFunc) error
	// Delete deletes object
	Delete(ctx context.Context) error
}

type arangoMemberModContext struct {
	cache   inspector.Inspector
	name    string
	timeout time.Duration
}

func (a arangoMemberModContext) withTimeout(ctx context.Context) (context.Context, func()) {
	if a.timeout != 0 {
		return context.WithTimeout(ctx, a.timeout)
	}

	return ctx, func() {}
}

func (a arangoMemberModContext) Delete(ctx context.Context) error {
	ctx, c := a.withTimeout(ctx)
	defer c()

	if err := a.cache.Client().Arango().DatabaseV1().ArangoMembers(a.cache.Namespace()).Delete(ctx, a.name, meta.DeleteOptions{}); err != nil {
		if api.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}

func (a arangoMemberModContext) Exists(ctx context.Context) bool {
	_, ok := a.cache.ArangoMember().V1().GetSimple(a.name)
	return ok
}

func (a arangoMemberModContext) Create(ctx context.Context, obj *api.ArangoMember) error {
	ctx, c := a.withTimeout(ctx)
	defer c()

	if obj.GetName() == "" {
		obj.Name = a.name
	} else if obj.GetName() != a.name {
		return errors.Newf("Name is invalid")
	}

	if obj.GetNamespace() == "" {
		obj.Namespace = a.cache.Namespace()
	} else if obj.GetNamespace() != a.cache.Namespace() {
		return errors.Newf("Namespace is invalid")
	}

	if _, err := a.cache.Client().Arango().DatabaseV1().ArangoMembers(obj.GetNamespace()).Create(ctx, obj, meta.CreateOptions{}); err != nil {
		return err
	}

	if err := a.cache.ArangoMember().Refresh(ctx); err != nil {
		return err
	}

	return nil
}

func (a arangoMemberModContext) Update(ctx context.Context, action ArangoMemberUpdateFunc) error {
	ctx, c := a.withTimeout(ctx)
	defer c()

	o, err := a.cache.ArangoMember().V1().Read().Get(ctx, a.name, meta.GetOptions{})
	if err != nil {
		return err
	}

	if action(o) {
		if _, err := a.cache.Client().Arango().DatabaseV1().ArangoMembers(a.cache.Namespace()).Update(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}

		if err := a.cache.ArangoMember().Refresh(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a arangoMemberModContext) UpdateStatus(ctx context.Context, action ArangoMemberStatusUpdateFunc) error {
	ctx, c := a.withTimeout(ctx)
	defer c()

	o, err := a.cache.ArangoMember().V1().Read().Get(ctx, a.name, meta.GetOptions{})
	if err != nil {
		return err
	}

	status := o.Status.DeepCopy()

	if action(o, status) {
		o.Status = *status
		if _, err := a.cache.Client().Arango().DatabaseV1().ArangoMembers(a.cache.Namespace()).UpdateStatus(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}

		if err := a.cache.ArangoMember().Refresh(ctx); err != nil {
			return err
		}
	}

	return nil
}
