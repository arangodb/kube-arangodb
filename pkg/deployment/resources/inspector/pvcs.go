//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
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

	q.v1 = newInspectorVersion[*core.PersistentVolumeClaimList, *core.PersistentVolumeClaim](ctx,
		inspectorConstants.PersistentVolumeClaimGRv1(),
		inspectorConstants.PersistentVolumeClaimGKv1(),
		i.client.Kubernetes().CoreV1().PersistentVolumeClaims(i.namespace),
		persistentvolumeclaim.List())

	i.persistentVolumeClaims = &q
	q.state = i
	q.last = time.Now()
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

	v1 *inspectorVersion[*core.PersistentVolumeClaim]
}

func (p *persistentVolumeClaimsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *persistentVolumeClaimsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, persistentVolumeClaimsInspectorLoaderObj)
}

func (p *persistentVolumeClaimsInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *persistentVolumeClaimsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.PersistentVolumeClaim()
}

func (p *persistentVolumeClaimsInspector) validate() error {
	if p == nil {
		return errors.Errorf("PersistentVolumeClaimInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *persistentVolumeClaimsInspector) V1() generic.Inspector[*core.PersistentVolumeClaim] {
	return p.v1
}
