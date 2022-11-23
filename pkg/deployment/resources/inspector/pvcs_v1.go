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
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
)

func (p *persistentVolumeClaimsInspector) V1() ins.Inspector {
	return p.v1
}

type persistentVolumeClaimsInspectorV1 struct {
	persistentVolumeClaimInspector *persistentVolumeClaimsInspector

	persistentVolumeClaims map[string]*core.PersistentVolumeClaim
	err                    error
}

func (p *persistentVolumeClaimsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("PersistentVolumeClaimsV1Inspector is nil")
	}

	if p.persistentVolumeClaimInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.persistentVolumeClaims == nil {
		return errors.Newf("PersistentVolumeClaims or err should be not nil")
	}

	if p.err != nil {
		return errors.Newf("PersistentVolumeClaims or err cannot be not nil together")
	}

	return nil
}

func (p *persistentVolumeClaimsInspectorV1) ListSimple() []*core.PersistentVolumeClaim {
	var r []*core.PersistentVolumeClaim
	for _, persistentVolumeClaim := range p.persistentVolumeClaims {
		r = append(r, persistentVolumeClaim)
	}

	return r
}

func (p *persistentVolumeClaimsInspectorV1) GetSimple(name string) (*core.PersistentVolumeClaim, bool) {
	persistentVolumeClaim, ok := p.persistentVolumeClaims[name]
	if !ok {
		return nil, false
	}

	return persistentVolumeClaim, true
}

func (p *persistentVolumeClaimsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, persistentVolumeClaim := range p.persistentVolumeClaims {
		if err := p.iteratePersistentVolumeClaim(persistentVolumeClaim, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *persistentVolumeClaimsInspectorV1) iteratePersistentVolumeClaim(persistentVolumeClaim *core.PersistentVolumeClaim, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(persistentVolumeClaim) {
			return nil
		}
	}

	return action(persistentVolumeClaim)
}

func (p *persistentVolumeClaimsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *persistentVolumeClaimsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.PersistentVolumeClaim, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.PersistentVolumeClaimGR(), name)
	} else {
		return s, nil
	}
}
