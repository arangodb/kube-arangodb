//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	requireRegisterInspectorLoader(persistentVolumesInspectorLoaderObj)
}

var persistentVolumesInspectorLoaderObj = persistentVolumesInspectorLoader{}

type persistentVolumesInspectorLoader struct {
}

func (p persistentVolumesInspectorLoader) Component() definitions.Component {
	return definitions.PersistentVolume
}

func (p persistentVolumesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q persistentVolumesInspector
	p.loadV1(ctx, i, &q)
	i.persistentVolumes = &q
	q.state = i
	q.last = time.Now()
}

func (p persistentVolumesInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *persistentVolumesInspector) {
	var z persistentVolumesInspectorV1

	z.persistentVolumeInspector = q

	z.persistentVolumes, z.err = p.getV1PersistentVolumes(ctx, i)

	q.v1 = &z
}

func (p persistentVolumesInspectorLoader) getV1PersistentVolumes(ctx context.Context, i *inspectorState) (map[string]*core.PersistentVolume, error) {
	objs, err := p.getV1PersistentVolumesList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.PersistentVolume, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p persistentVolumesInspectorLoader) getV1PersistentVolumesList(ctx context.Context, i *inspectorState) ([]*core.PersistentVolume, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().PersistentVolumes().List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.PersistentVolume, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1PersistentVolumesListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p persistentVolumesInspectorLoader) getV1PersistentVolumesListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.PersistentVolume, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().PersistentVolumes().List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p persistentVolumesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p persistentVolumesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.persistentVolumes != nil {
		if !override {
			return
		}
	}

	to.persistentVolumes = from.persistentVolumes
	to.persistentVolumes.state = to
}

func (p persistentVolumesInspectorLoader) Name() string {
	return "persistentVolumes"
}

type persistentVolumesInspector struct {
	state *inspectorState

	last time.Time

	v1 *persistentVolumesInspectorV1
}

func (p *persistentVolumesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *persistentVolumesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, persistentVolumesInspectorLoaderObj)
}

func (p *persistentVolumesInspector) Version() version.Version {
	return version.V1
}

func (p *persistentVolumesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.PersistentVolume()
}

func (p *persistentVolumesInspector) validate() error {
	if p == nil {
		return errors.Newf("PersistentVolumeInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
