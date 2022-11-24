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

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
)

func (p *podsInspector) V1() ins.Inspector {

	return p.v1
}

type podsInspectorV1 struct {
	podInspector *podsInspector

	pods map[string]*core.Pod
	err  error
}

func (p *podsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("PodsV1Inspector is nil")
	}

	if p.podInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.pods == nil {
		return errors.Newf("Pods or err should be not nil")
	}

	if p.err != nil {
		return errors.Newf("Pods or err cannot be not nil together")
	}

	return nil
}

func (p *podsInspectorV1) ListSimple() []*core.Pod {
	var r []*core.Pod
	for _, pod := range p.pods {
		r = append(r, pod)
	}

	return r
}

func (p *podsInspectorV1) GetSimple(name string) (*core.Pod, bool) {
	pod, ok := p.pods[name]
	if !ok {
		return nil, false
	}

	return pod, true
}

func (p *podsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, pod := range p.pods {
		if err := p.iteratePod(pod, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *podsInspectorV1) iteratePod(pod *core.Pod, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(pod) {
			return nil
		}
	}

	return action(pod)
}

func (p *podsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *podsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Pod, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.PodGR(), name)
	} else {
		return s, nil
	}
}
