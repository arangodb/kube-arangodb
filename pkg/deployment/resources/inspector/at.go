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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(arangoTasksInspectorLoaderObj)
}

var arangoTasksInspectorLoaderObj = arangoTasksInspectorLoader{}

type arangoTasksInspectorLoader struct {
}

func (p arangoTasksInspectorLoader) Component() definitions.Component {
	return definitions.ArangoTask
}

func (p arangoTasksInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoTasksInspector

	q.v1 = newInspectorVersion[*api.ArangoTaskList, *api.ArangoTask](ctx,
		inspectorConstants.ArangoTaskGRv1(),
		inspectorConstants.ArangoTaskGKv1(),
		i.client.Arango().DatabaseV1().ArangoTasks(i.namespace),
		arangotask.List())

	i.arangoTasks = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoTasksInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoTasksInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoTasks != nil {
		if !override {
			return
		}
	}

	to.arangoTasks = from.arangoTasks
	to.arangoTasks.state = to
}

func (p arangoTasksInspectorLoader) Name() string {
	return "arangoTasks"
}

type arangoTasksInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*api.ArangoTask]
}

func (p *arangoTasksInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoTasksInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoTasksInspectorLoaderObj)
}

func (p *arangoTasksInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *arangoTasksInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoTask()
}

func (p *arangoTasksInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoTaskInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *arangoTasksInspector) V1() (generic.Inspector[*api.ArangoTask], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
