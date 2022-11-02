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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(arangoMembersInspectorLoaderObj)
}

var arangoMembersInspectorLoaderObj = arangoMembersInspectorLoader{}

type arangoMembersInspectorLoader struct {
}

func (p arangoMembersInspectorLoader) Component() definitions.Component {
	return definitions.ArangoMember
}

func (p arangoMembersInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoMembersInspector
	p.loadV1(ctx, i, &q)
	i.arangoMembers = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoMembersInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *arangoMembersInspector) {
	var z arangoMembersInspectorV1

	z.arangoMemberInspector = q

	z.arangoMembers, z.err = p.getV1ArangoMembers(ctx, i)

	q.v1 = &z
}

func (p arangoMembersInspectorLoader) getV1ArangoMembers(ctx context.Context, i *inspectorState) (map[string]*api.ArangoMember, error) {
	objs, err := p.getV1ArangoMembersList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*api.ArangoMember, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p arangoMembersInspectorLoader) getV1ArangoMembersList(ctx context.Context, i *inspectorState) ([]*api.ArangoMember, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoMembers(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*api.ArangoMember, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ArangoMembersListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p arangoMembersInspectorLoader) getV1ArangoMembersListRequest(ctx context.Context, i *inspectorState, cont string) ([]api.ArangoMember, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoMembers(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p arangoMembersInspectorLoader) Verify(i *inspectorState) error {
	if err := i.arangoMembers.v1.err; err != nil {
		return err
	}

	return nil
}

func (p arangoMembersInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoMembers != nil {
		if !override {
			return
		}
	}

	to.arangoMembers = from.arangoMembers
	to.arangoMembers.state = to
}

func (p arangoMembersInspectorLoader) Name() string {
	return "arangoMembers"
}

type arangoMembersInspector struct {
	state *inspectorState

	last time.Time

	v1 *arangoMembersInspectorV1
}

func (p *arangoMembersInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoMembersInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoMembersInspectorLoaderObj)
}

func (p *arangoMembersInspector) Version() version.Version {
	return version.V1
}

func (p *arangoMembersInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoMember()
}

func (p *arangoMembersInspector) validate() error {
	if p == nil {
		return errors.Newf("ArangoMemberInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
