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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoprofile/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoProfilesInspector) V1Beta1() (ins.Inspector, error) {
	if p.v1beta1.err != nil {
		return nil, p.v1beta1.err
	}

	return p.v1beta1, nil
}

type arangoProfilesInspectorV1Beta1 struct {
	arangoProfileInspector *arangoProfilesInspector

	arangoProfiles map[string]*schedulerApi.ArangoProfile
	err            error
}

func (p *arangoProfilesInspectorV1Beta1) Filter(filters ...ins.Filter) []*schedulerApi.ArangoProfile {
	z := p.ListSimple()

	r := make([]*schedulerApi.ArangoProfile, 0, len(z))

	for _, o := range z {
		if !ins.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *arangoProfilesInspectorV1Beta1) validate() error {
	if p == nil {
		return errors.Errorf("ArangoProfilesV1AlphaInspector is nil")
	}

	if p.arangoProfileInspector == nil {
		return errors.Errorf("Parent is nil")
	}

	if p.arangoProfiles == nil && p.err == nil {
		return errors.Errorf("ArangoProfiles or err should be not nil")
	}

	if p.arangoProfiles != nil && p.err != nil {
		return errors.Errorf("ArangoProfiles or err cannot be not nil together")
	}

	return nil
}

func (p *arangoProfilesInspectorV1Beta1) ListSimple() []*schedulerApi.ArangoProfile {
	var r []*schedulerApi.ArangoProfile
	for _, arangoProfile := range p.arangoProfiles {
		r = append(r, arangoProfile)
	}

	return r
}

func (p *arangoProfilesInspectorV1Beta1) GetSimple(name string) (*schedulerApi.ArangoProfile, bool) {
	arangoProfile, ok := p.arangoProfiles[name]
	if !ok {
		return nil, false
	}

	return arangoProfile, true
}

func (p *arangoProfilesInspectorV1Beta1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, arangoProfile := range p.arangoProfiles {
		if err := p.iterateArangoProfile(arangoProfile, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *arangoProfilesInspectorV1Beta1) iterateArangoProfile(arangoProfile *schedulerApi.ArangoProfile, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(arangoProfile) {
			return nil
		}
	}

	return action(arangoProfile)
}

func (p *arangoProfilesInspectorV1Beta1) Read() ins.ReadInterface {
	return p
}

func (p *arangoProfilesInspectorV1Beta1) Get(ctx context.Context, name string, opts meta.GetOptions) (*schedulerApi.ArangoProfile, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ArangoProfileGR(), name)
	} else {
		return s, nil
	}
}
