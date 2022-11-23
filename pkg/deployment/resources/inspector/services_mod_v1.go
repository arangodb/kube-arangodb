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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
)

func (p servicesMod) V1() servicev1.ModInterface {
	return servicesModV1(p)
}

type servicesModV1 struct {
	i *inspectorState
}

func (p servicesModV1) client() typedCore.ServiceInterface {
	return p.i.Client().Kubernetes().CoreV1().Services(p.i.Namespace())
}

func (p servicesModV1) Create(ctx context.Context, service *core.Service, opts meta.CreateOptions) (*core.Service, error) {
	if service, err := p.client().Create(ctx, service, opts); err != nil {
		return service, err
	} else {
		p.i.GetThrottles().Service().Invalidate()
		return service, err
	}
}

func (p servicesModV1) Update(ctx context.Context, service *core.Service, opts meta.UpdateOptions) (*core.Service, error) {
	if service, err := p.client().Update(ctx, service, opts); err != nil {
		return service, err
	} else {
		p.i.GetThrottles().Service().Invalidate()
		return service, err
	}
}

func (p servicesModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.Service, err error) {
	if service, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return service, err
	} else {
		p.i.GetThrottles().Service().Invalidate()
		return service, err
	}
}

func (p servicesModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().Service().Invalidate()
		return err
	}
}
