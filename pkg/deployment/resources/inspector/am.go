//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
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

	q.v1 = newInspectorVersion[*api.ArangoMemberList, *api.ArangoMember](ctx,
		constants.ArangoMemberGRv1(),
		constants.ArangoMemberGKv1(),
		i.client.Arango().DatabaseV1().ArangoMembers(i.namespace),
		arangomember.List())

	i.arangoMembers = &q
	q.state = i
	q.last = time.Now()
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

	v1 *inspectorVersion[*api.ArangoMember]
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
		return errors.Errorf("ArangoMemberInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *arangoMembersInspector) V1() generic.Inspector[*api.ArangoMember] {
	return p.v1
}
