//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
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
	p.loadV1Beta1(ctx, i, &q)
	i.arangoProfiles = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoProfilesInspectorLoader) loadV1Beta1(ctx context.Context, i *inspectorState, q *arangoProfilesInspector) {
	var z arangoProfilesInspectorV1Beta1

	z.arangoProfileInspector = q

	z.arangoProfiles, z.err = p.getV1ArangoProfiles(ctx, i)

	q.v1beta1 = &z
}

func (p arangoProfilesInspectorLoader) getV1ArangoProfiles(ctx context.Context, i *inspectorState) (map[string]*schedulerApi.ArangoProfile, error) {
	objs, err := p.getV1ArangoProfilesList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*schedulerApi.ArangoProfile, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p arangoProfilesInspectorLoader) getV1ArangoProfilesList(ctx context.Context, i *inspectorState) ([]*schedulerApi.ArangoProfile, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().SchedulerV1beta1().ArangoProfiles(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*schedulerApi.ArangoProfile, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ArangoProfilesListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p arangoProfilesInspectorLoader) getV1ArangoProfilesListRequest(ctx context.Context, i *inspectorState, cont string) ([]schedulerApi.ArangoProfile, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().SchedulerV1beta1().ArangoProfiles(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
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

	v1beta1 *arangoProfilesInspectorV1Beta1
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
