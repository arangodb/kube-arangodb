//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoplatformstorage"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(arangoPlatformStoragesInspectorLoaderObj)
}

var arangoPlatformStoragesInspectorLoaderObj = arangoPlatformStoragesInspectorLoader{}

type arangoPlatformStoragesInspectorLoader struct {
}

func (p arangoPlatformStoragesInspectorLoader) Component() definitions.Component {
	return definitions.ArangoPlatformStorage
}

func (p arangoPlatformStoragesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoPlatformStoragesInspector

	q.v1beta1 = newInspectorVersion[*platformApi.ArangoPlatformStorageList, *platformApi.ArangoPlatformStorage](ctx,
		inspectorConstants.ArangoPlatformStorageGRv1Beta1(),
		inspectorConstants.ArangoPlatformStorageGKv1Beta1(),
		i.client.Arango().PlatformV1beta1().ArangoPlatformStorages(i.namespace),
		arangoplatformstorage.List())

	i.arangoPlatformStorages = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoPlatformStoragesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoPlatformStoragesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoPlatformStorages != nil {
		if !override {
			return
		}
	}

	to.arangoPlatformStorages = from.arangoPlatformStorages
	to.arangoPlatformStorages.state = to
}

func (p arangoPlatformStoragesInspectorLoader) Name() string {
	return "arangoPlatformStorages"
}

type arangoPlatformStoragesInspector struct {
	state *inspectorState

	last time.Time

	v1beta1 *inspectorVersion[*platformApi.ArangoPlatformStorage]
}

func (p *arangoPlatformStoragesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoPlatformStoragesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoPlatformStoragesInspectorLoaderObj)
}

func (p *arangoPlatformStoragesInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1Beta1
}

func (p *arangoPlatformStoragesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoPlatformStorage()
}

func (p *arangoPlatformStoragesInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoPlatformStorageInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1beta1.validate()
}

func (p *arangoPlatformStoragesInspector) V1Beta1() (generic.Inspector[*platformApi.ArangoPlatformStorage], error) {
	if p.v1beta1.err != nil {
		return nil, p.v1beta1.err
	}

	return p.v1beta1, nil
}
