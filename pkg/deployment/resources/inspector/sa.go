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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(serviceAccountsInspectorLoaderObj)
}

var serviceAccountsInspectorLoaderObj = serviceAccountsInspectorLoader{}

type serviceAccountsInspectorLoader struct {
}

func (p serviceAccountsInspectorLoader) Component() definitions.Component {
	return definitions.ServiceAccount
}

func (p serviceAccountsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q serviceAccountsInspector

	q.v1 = newInspectorVersion[*core.ServiceAccountList, *core.ServiceAccount](ctx,
		constants.ServiceAccountGRv1(),
		constants.ServiceAccountGKv1(),
		i.client.Kubernetes().CoreV1().ServiceAccounts(i.namespace),
		serviceaccount.List())

	i.serviceAccounts = &q
	q.state = i
	q.last = time.Now()
}

func (p serviceAccountsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.serviceAccounts.v1.err; err != nil {
		return err
	}

	return nil
}

func (p serviceAccountsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.serviceAccounts != nil {
		if !override {
			return
		}
	}

	to.serviceAccounts = from.serviceAccounts
	to.serviceAccounts.state = to
}

func (p serviceAccountsInspectorLoader) Name() string {
	return "serviceAccounts"
}

type serviceAccountsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.ServiceAccount]
}

func (p *serviceAccountsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *serviceAccountsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, serviceAccountsInspectorLoaderObj)
}

func (p *serviceAccountsInspector) Version() version.Version {
	return version.V1
}

func (p *serviceAccountsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ServiceAccount()
}

func (p *serviceAccountsInspector) validate() error {
	if p == nil {
		return errors.Errorf("ServiceAccountInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *serviceAccountsInspector) V1() generic.Inspector[*core.ServiceAccount] {
	return p.v1
}
