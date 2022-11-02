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

package inspector

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(secretsInspectorLoaderObj)
}

var secretsInspectorLoaderObj = secretsInspectorLoader{}

type secretsInspectorLoader struct {
}

func (p secretsInspectorLoader) Component() definitions.Component {
	return definitions.Secret
}

func (p secretsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q secretsInspector
	p.loadV1(ctx, i, &q)
	i.secrets = &q
	q.state = i
	q.last = time.Now()
}

func (p secretsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *secretsInspector) {
	var z secretsInspectorV1

	z.secretInspector = q

	z.secrets, z.err = p.getV1Secrets(ctx, i)

	q.v1 = &z
}

func (p secretsInspectorLoader) getV1Secrets(ctx context.Context, i *inspectorState) (map[string]*core.Secret, error) {
	objs, err := p.getV1SecretsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.Secret, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p secretsInspectorLoader) getV1SecretsList(ctx context.Context, i *inspectorState) ([]*core.Secret, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Secrets(i.namespace).List(ctxChild, meta.ListOptions{
		Limit: globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
	})

	if err != nil {
		return nil, err
	}

	items := obj.Items
	cont := obj.Continue
	var s = int64(len(items))

	if z := obj.RemainingItemCount; z != nil {
		s += *z
	}

	ptrs := make([]*core.Secret, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1SecretsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p secretsInspectorLoader) getV1SecretsListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.Secret, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Secrets(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p secretsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.secrets.v1.err; err != nil {
		return err
	}

	return nil
}

func (p secretsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.secrets != nil {
		if !override {
			return
		}
	}

	to.secrets = from.secrets
	to.secrets.state = to
}

func (p secretsInspectorLoader) Name() string {
	return "secrets"
}

type secretsInspector struct {
	state *inspectorState

	last time.Time

	v1 *secretsInspectorV1
}

func (p *secretsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *secretsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, secretsInspectorLoaderObj)
}

func (p *secretsInspector) Version() version.Version {
	return version.V1
}

func (p *secretsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Secret()
}

func (p *secretsInspector) validate() error {
	if p == nil {
		return errors.Newf("SecretInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
