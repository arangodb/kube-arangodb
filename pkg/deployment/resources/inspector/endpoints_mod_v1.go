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

	endpointsv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (p endpointsMod) V1() endpointsv1.ModInterface {
	return endpointsModV1(p)
}

type endpointsModV1 struct {
	i *inspectorState
}

func (p endpointsModV1) client() typedCore.EndpointsInterface {
	return p.i.Client().Kubernetes().CoreV1().Endpoints(p.i.Namespace())
}

func (p endpointsModV1) Create(ctx context.Context, endpoint *core.Endpoints, opts meta.CreateOptions) (*core.Endpoints, error) {
	if endpoint, err := p.client().Create(ctx, endpoint, opts); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().Endpoints().Invalidate()
		return endpoint, err
	}
}

func (p endpointsModV1) Update(ctx context.Context, endpoint *core.Endpoints, opts meta.UpdateOptions) (*core.Endpoints, error) {
	if endpoint, err := p.client().Update(ctx, endpoint, opts); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().Endpoints().Invalidate()
		return endpoint, err
	}
}

func (p endpointsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.Endpoints, err error) {
	if endpoint, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().Endpoints().Invalidate()
		return endpoint, err
	}
}

func (p endpointsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().Endpoints().Invalidate()
		return err
	}
}
