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

	serviceAccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (p serviceAccountsMod) V1() serviceAccountv1.ModInterface {
	return serviceAccountsModV1(p)
}

type serviceAccountsModV1 struct {
	i *inspectorState
}

func (p serviceAccountsModV1) client() typedCore.ServiceAccountInterface {
	return p.i.Client().Kubernetes().CoreV1().ServiceAccounts(p.i.Namespace())
}

func (p serviceAccountsModV1) Create(ctx context.Context, serviceAccount *core.ServiceAccount, opts meta.CreateOptions) (*core.ServiceAccount, error) {
	if serviceAccount, err := p.client().Create(ctx, serviceAccount, opts); err != nil {
		return serviceAccount, err
	} else {
		p.i.GetThrottles().ServiceAccount().Invalidate()
		return serviceAccount, err
	}
}

func (p serviceAccountsModV1) Update(ctx context.Context, serviceAccount *core.ServiceAccount, opts meta.UpdateOptions) (*core.ServiceAccount, error) {
	if serviceAccount, err := p.client().Update(ctx, serviceAccount, opts); err != nil {
		return serviceAccount, err
	} else {
		p.i.GetThrottles().ServiceAccount().Invalidate()
		return serviceAccount, err
	}
}

func (p serviceAccountsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.ServiceAccount, err error) {
	if serviceAccount, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return serviceAccount, err
	} else {
		p.i.GetThrottles().ServiceAccount().Invalidate()
		return serviceAccount, err
	}
}

func (p serviceAccountsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().ServiceAccount().Invalidate()
		return err
	}
}
