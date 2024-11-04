//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
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

	q.v1 = newInspectorVersion[*core.SecretList, *core.Secret](ctx,
		constants.SecretGRv1(),
		constants.SecretGKv1(),
		i.client.Kubernetes().CoreV1().Secrets(i.namespace),
		secret.List())

	i.secrets = &q
	q.state = i
	q.last = time.Now()
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

	v1 *inspectorVersion[*core.Secret]
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
		return errors.Errorf("SecretInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *secretsInspector) V1() generic.Inspector[*core.Secret] {
	return p.v1
}
