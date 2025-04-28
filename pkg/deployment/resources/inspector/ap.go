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

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoprofile"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(arangoProfilesInspectorLoaderObj)
}

var arangoProfilesInspectorLoaderObj = arangoProfilesInspectorLoader{}

type arangoProfilesInspectorLoader struct {
}

func (p arangoProfilesInspectorLoader) Component() definitions.Component {
	return definitions.ArangoProfile
}

func (p arangoProfilesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoProfilesInspector

	q.v1beta1 = newInspectorVersion[*schedulerApi.ArangoProfileList, *schedulerApi.ArangoProfile](ctx,
		constants.ArangoProfileGRv1Beta1(),
		constants.ArangoProfileGKv1Beta1(),
		i.client.Arango().SchedulerV1beta1().ArangoProfiles(i.namespace),
		arangoprofile.List())

	i.arangoProfiles = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoProfilesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoProfilesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoProfiles != nil {
		if !override {
			return
		}
	}

	to.arangoProfiles = from.arangoProfiles
	to.arangoProfiles.state = to
}

func (p arangoProfilesInspectorLoader) Name() string {
	return "arangoProfiles"
}

type arangoProfilesInspector struct {
	state *inspectorState

	last time.Time

	v1beta1 *inspectorVersion[*schedulerApi.ArangoProfile]
}

func (p *arangoProfilesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoProfilesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoProfilesInspectorLoaderObj)
}

func (p *arangoProfilesInspector) Version() version.Version {
	return version.V1
}

func (p *arangoProfilesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoProfile()
}

func (p *arangoProfilesInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoProfileInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1beta1.validate()
}

func (p *arangoProfilesInspector) V1Beta1() (generic.Inspector[*schedulerApi.ArangoProfile], error) {
	if p.v1beta1.err != nil {
		return nil, p.v1beta1.err
	}

	return p.v1beta1, nil
}
