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

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *configMapsInspector) V1() ins.Inspector {
	return p.v1
}

type configMapsInspectorV1 struct {
	configMapInspector *configMapsInspector

	configMaps map[string]*core.ConfigMap
	err        error
}

func (p *configMapsInspectorV1) validate() error {
	if p == nil {
		return errors.Errorf("ConfigMapsV1Inspector is nil")
	}

	if p.configMapInspector == nil {
		return errors.Errorf("Parent is nil")
	}

	if p.configMaps == nil {
		return errors.Errorf("ConfigMaps or err should be not nil")
	}

	if p.err != nil {
		return errors.Errorf("ConfigMaps or err cannot be not nil together")
	}

	return nil
}

func (p *configMapsInspectorV1) ListSimple() []*core.ConfigMap {
	var r []*core.ConfigMap
	for _, configMap := range p.configMaps {
		r = append(r, configMap)
	}

	return r
}

func (p *configMapsInspectorV1) GetSimple(name string) (*core.ConfigMap, bool) {
	configMap, ok := p.configMaps[name]
	if !ok {
		return nil, false
	}

	return configMap, true
}

func (p *configMapsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, configMap := range p.configMaps {
		if err := p.iterateConfigMap(configMap, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *configMapsInspectorV1) iterateConfigMap(configMap *core.ConfigMap, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(configMap) {
			return nil
		}
	}

	return action(configMap)
}

func (p *configMapsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *configMapsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.ConfigMap, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ConfigMapGR(), name)
	} else {
		return s, nil
	}
}
