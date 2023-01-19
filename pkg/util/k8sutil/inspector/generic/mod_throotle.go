//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package generic

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

type ThrottleGetter func() throttle.Components
type ClientStatusGetter[S meta.Object] func() ModStatusClient[S]

func NewModThrottle[S meta.Object](component definitions.Component, componentGetter ThrottleGetter, clientStatus ClientStatusGetter[S]) ModStatusClient[S] {
	return modThrottle[S]{
		component:       component,
		componentGetter: componentGetter,
		clientStatus:    clientStatus,
	}
}

type modThrottle[S meta.Object] struct {
	component       definitions.Component
	componentGetter ThrottleGetter
	clientStatus    ClientStatusGetter[S]
}

func (m modThrottle[S]) getClient() ModStatusClient[S] {
	return m.clientStatus()
}

func (m modThrottle[S]) throttle() {
	m.componentGetter().Invalidate(m.component)
}

func (m modThrottle[S]) Create(ctx context.Context, obj S, opts meta.CreateOptions) (S, error) {
	if obj, err := m.getClient().Create(ctx, obj, opts); err != nil {
		return obj, err
	} else {
		m.throttle()
		return obj, err
	}
}

func (m modThrottle[S]) Update(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	if obj, err := m.getClient().Update(ctx, obj, opts); err != nil {
		return obj, err
	} else {
		m.throttle()
		return obj, err
	}
}

func (m modThrottle[S]) UpdateStatus(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	if obj, err := m.getClient().UpdateStatus(ctx, obj, opts); err != nil {
		return obj, err
	} else {
		m.throttle()
		return obj, err
	}
}

func (m modThrottle[S]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result S, err error) {
	if obj, err := m.getClient().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return obj, err
	} else {
		m.throttle()
		return obj, err
	}
}

func (m modThrottle[S]) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := m.getClient().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		m.throttle()
		return err
	}
}
