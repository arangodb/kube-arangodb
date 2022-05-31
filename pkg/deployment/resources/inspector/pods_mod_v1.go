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

	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (p podsMod) V1() podv1.ModInterface {
	return podsModV1(p)
}

type podsModV1 struct {
	i *inspectorState
}

func (p podsModV1) client() typedCore.PodInterface {
	return p.i.Client().Kubernetes().CoreV1().Pods(p.i.Namespace())
}

func (p podsModV1) Create(ctx context.Context, pod *core.Pod, opts meta.CreateOptions) (*core.Pod, error) {
	if pod, err := p.client().Create(ctx, pod, opts); err != nil {
		return pod, err
	} else {
		p.i.GetThrottles().Pod().Invalidate()
		return pod, err
	}
}

func (p podsModV1) Update(ctx context.Context, pod *core.Pod, opts meta.UpdateOptions) (*core.Pod, error) {
	if pod, err := p.client().Update(ctx, pod, opts); err != nil {
		return pod, err
	} else {
		p.i.GetThrottles().Pod().Invalidate()
		return pod, err
	}
}

func (p podsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.Pod, err error) {
	if pod, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return pod, err
	} else {
		p.i.GetThrottles().Pod().Invalidate()
		return pod, err
	}
}

func (p podsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().Pod().Invalidate()
		return err
	}
}
