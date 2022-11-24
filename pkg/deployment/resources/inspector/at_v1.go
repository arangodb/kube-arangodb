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
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoTasksInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type arangoTasksInspectorV1 struct {
	arangoTaskInspector *arangoTasksInspector

	arangoTasks map[string]*api.ArangoTask
	err         error
}

func (p *arangoTasksInspectorV1) Filter(filters ...ins.Filter) []*api.ArangoTask {
	z := p.ListSimple()

	r := make([]*api.ArangoTask, 0, len(z))

	for _, o := range z {
		if !ins.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *arangoTasksInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("ArangoTasksV1Inspector is nil")
	}

	if p.arangoTaskInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.arangoTasks == nil && p.err == nil {
		return errors.Newf("ArangoTasks or err should be not nil")
	}

	if p.arangoTasks != nil && p.err != nil {
		return errors.Newf("ArangoTasks or err cannot be not nil together")
	}

	return nil
}

func (p *arangoTasksInspectorV1) ListSimple() []*api.ArangoTask {
	var r []*api.ArangoTask
	for _, arangoTask := range p.arangoTasks {
		r = append(r, arangoTask)
	}

	return r
}

func (p *arangoTasksInspectorV1) GetSimple(name string) (*api.ArangoTask, bool) {
	arangoTask, ok := p.arangoTasks[name]
	if !ok {
		return nil, false
	}

	return arangoTask, true
}

func (p *arangoTasksInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, arangoTask := range p.arangoTasks {
		if err := p.iterateArangoTask(arangoTask, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *arangoTasksInspectorV1) iterateArangoTask(arangoTask *api.ArangoTask, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(arangoTask) {
			return nil
		}
	}

	return action(arangoTask)
}

func (p *arangoTasksInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *arangoTasksInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*api.ArangoTask, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ArangoTaskGR(), name)
	} else {
		return s, nil
	}
}
