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
	requireRegisterInspectorLoader(persistentVolumeClaimsInspectorLoaderObj)
}

var persistentVolumeClaimsInspectorLoaderObj = persistentVolumeClaimsInspectorLoader{}

type persistentVolumeClaimsInspectorLoader struct {
}

func (p persistentVolumeClaimsInspectorLoader) Component() definitions.Component {
	return definitions.PersistentVolumeClaim
}

func (p persistentVolumeClaimsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q persistentVolumeClaimsInspector
	p.loadV1(ctx, i, &q)
	i.persistentVolumeClaims = &q
	q.state = i
	q.last = time.Now()
}

func (p persistentVolumeClaimsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *persistentVolumeClaimsInspector) {
	var z persistentVolumeClaimsInspectorV1

	z.persistentVolumeClaimInspector = q

	z.persistentVolumeClaims, z.err = p.getV1PersistentVolumeClaims(ctx, i)

	q.v1 = &z
}

func (p persistentVolumeClaimsInspectorLoader) getV1PersistentVolumeClaims(ctx context.Context, i *inspectorState) (map[string]*core.PersistentVolumeClaim, error) {
	objs, err := p.getV1PersistentVolumeClaimsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.PersistentVolumeClaim, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p persistentVolumeClaimsInspectorLoader) getV1PersistentVolumeClaimsList(ctx context.Context, i *inspectorState) ([]*core.PersistentVolumeClaim, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().PersistentVolumeClaims(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.PersistentVolumeClaim, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1PersistentVolumeClaimsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p persistentVolumeClaimsInspectorLoader) getV1PersistentVolumeClaimsListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.PersistentVolumeClaim, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().PersistentVolumeClaims(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p persistentVolumeClaimsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.persistentVolumeClaims.v1.err; err != nil {
		return err
	}

	return nil
}

func (p persistentVolumeClaimsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.persistentVolumeClaims != nil {
		if !override {
			return
		}
	}

	to.persistentVolumeClaims = from.persistentVolumeClaims
	to.persistentVolumeClaims.state = to
}

func (p persistentVolumeClaimsInspectorLoader) Name() string {
	return "persistentVolumeClaims"
}

type persistentVolumeClaimsInspector struct {
	state *inspectorState

	last time.Time

	v1 *persistentVolumeClaimsInspectorV1
}

func (p *persistentVolumeClaimsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *persistentVolumeClaimsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, persistentVolumeClaimsInspectorLoaderObj)
}

func (p *persistentVolumeClaimsInspector) Version() version.Version {
	return version.V1
}

func (p *persistentVolumeClaimsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.PersistentVolumeClaim()
}

func (p *persistentVolumeClaimsInspector) validate() error {
	if p == nil {
		return errors.Newf("PersistentVolumeClaimInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
