//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolume"
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

	q.v1 = newInspectorVersion[*core.PersistentVolumeList, *core.PersistentVolume](ctx,
		constants.PersistentVolumeGRv1(),
		constants.PersistentVolumeGKv1(),
		i.client.Kubernetes().CoreV1().PersistentVolumes(),
		persistentvolume.List())

	i.persistentVolumes = &q
	q.state = i
	q.last = time.Now()
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

	v1 *inspectorVersion[*core.PersistentVolume]
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
		return errors.Errorf("PersistentVolumeInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *persistentVolumesInspector) V1() (generic.Inspector[*core.PersistentVolume], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
