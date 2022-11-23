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
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
)

func (p *secretsInspector) V1() ins.Inspector {
	return p.v1
}

type secretsInspectorV1 struct {
	secretInspector *secretsInspector

	secrets map[string]*core.Secret
	err     error
}

func (p *secretsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("SecretsV1Inspector is nil")
	}

	if p.secretInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.secrets == nil {
		return errors.Newf("Secrets or err should be not nil")
	}

	if p.err != nil {
		return errors.Newf("Secrets or err cannot be not nil together")
	}

	return nil
}

func (p *secretsInspectorV1) ListSimple() []*core.Secret {
	var r []*core.Secret
	for _, secret := range p.secrets {
		r = append(r, secret)
	}

	return r
}

func (p *secretsInspectorV1) GetSimple(name string) (*core.Secret, bool) {
	secret, ok := p.secrets[name]
	if !ok {
		return nil, false
	}

	return secret, true
}

func (p *secretsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, secret := range p.secrets {
		if err := p.iterateSecret(secret, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *secretsInspectorV1) iterateSecret(secret *core.Secret, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(secret) {
			return nil
		}
	}

	return action(secret)
}

func (p *secretsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *secretsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Secret, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.SecretGR(), name)
	} else {
		return s, nil
	}
}
