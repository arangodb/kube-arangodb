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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoMembersInspector) V1() ins.Inspector {
	return p.v1
}

type arangoMembersInspectorV1 struct {
	arangoMemberInspector *arangoMembersInspector

	arangoMembers map[string]*api.ArangoMember
	err           error
}

func (p *arangoMembersInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("ArangoMembersV1Inspector is nil")
	}

	if p.arangoMemberInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.arangoMembers == nil {
		return errors.Newf("ArangoMembers or err should be not nil")
	}

	if p.err != nil {
		return errors.Newf("ArangoMembers or err cannot be not nil together")
	}

	return nil
}

func (p *arangoMembersInspectorV1) ArangoMembers() []*api.ArangoMember {
	var r []*api.ArangoMember
	for _, arangoMember := range p.arangoMembers {
		r = append(r, arangoMember)
	}

	return r
}

func (p *arangoMembersInspectorV1) GetSimple(name string) (*api.ArangoMember, bool) {
	arangoMember, ok := p.arangoMembers[name]
	if !ok {
		return nil, false
	}

	return arangoMember, true
}

func (p *arangoMembersInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, arangoMember := range p.arangoMembers {
		if err := p.iterateArangoMember(arangoMember, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *arangoMembersInspectorV1) iterateArangoMember(arangoMember *api.ArangoMember, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(arangoMember) {
			return nil
		}
	}

	return action(arangoMember)
}

func (p *arangoMembersInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *arangoMembersInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*api.ArangoMember, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ArangoMemberGR(), name)
	} else {
		return s, nil
	}
}
