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

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
)

func (p *serviceAccountsInspector) V1() ins.Inspector {
	return p.v1
}

type serviceAccountsInspectorV1 struct {
	serviceAccountInspector *serviceAccountsInspector

	serviceAccounts map[string]*core.ServiceAccount
	err             error
}

func (p *serviceAccountsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("ServiceAccountsV1Inspector is nil")
	}

	if p.serviceAccountInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.serviceAccounts == nil {
		return errors.Newf("ServiceAccounts or err should be not nil")
	}

	if p.err != nil {
		return errors.Newf("ServiceAccounts or err cannot be not nil together")
	}

	return nil
}

func (p *serviceAccountsInspectorV1) ServiceAccounts() []*core.ServiceAccount {
	var r []*core.ServiceAccount
	for _, serviceAccount := range p.serviceAccounts {
		r = append(r, serviceAccount)
	}

	return r
}

func (p *serviceAccountsInspectorV1) GetSimple(name string) (*core.ServiceAccount, bool) {
	serviceAccount, ok := p.serviceAccounts[name]
	if !ok {
		return nil, false
	}

	return serviceAccount, true
}

func (p *serviceAccountsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, serviceAccount := range p.serviceAccounts {
		if err := p.iterateServiceAccount(serviceAccount, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *serviceAccountsInspectorV1) iterateServiceAccount(serviceAccount *core.ServiceAccount, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(serviceAccount) {
			return nil
		}
	}

	return action(serviceAccount)
}

func (p *serviceAccountsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *serviceAccountsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.ServiceAccount, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ServiceAccountGR(), name)
	} else {
		return s, nil
	}
}
