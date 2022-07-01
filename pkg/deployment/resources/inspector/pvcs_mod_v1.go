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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	persistentVolumeClaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
)

func (p persistentVolumeClaimsMod) V1() persistentVolumeClaimv1.ModInterface {
	return persistentVolumeClaimsModV1(p)
}

type persistentVolumeClaimsModV1 struct {
	i *inspectorState
}

func (p persistentVolumeClaimsModV1) client() typedCore.PersistentVolumeClaimInterface {
	return p.i.Client().Kubernetes().CoreV1().PersistentVolumeClaims(p.i.Namespace())
}

func (p persistentVolumeClaimsModV1) Create(ctx context.Context, persistentVolumeClaim *core.PersistentVolumeClaim, opts meta.CreateOptions) (*core.PersistentVolumeClaim, error) {
	if persistentVolumeClaim, err := p.client().Create(ctx, persistentVolumeClaim, opts); err != nil {
		return persistentVolumeClaim, err
	} else {
		p.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return persistentVolumeClaim, err
	}
}

func (p persistentVolumeClaimsModV1) Update(ctx context.Context, persistentVolumeClaim *core.PersistentVolumeClaim, opts meta.UpdateOptions) (*core.PersistentVolumeClaim, error) {
	if persistentVolumeClaim, err := p.client().Update(ctx, persistentVolumeClaim, opts); err != nil {
		return persistentVolumeClaim, err
	} else {
		p.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return persistentVolumeClaim, err
	}
}

func (p persistentVolumeClaimsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.PersistentVolumeClaim, err error) {
	if persistentVolumeClaim, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return persistentVolumeClaim, err
	} else {
		p.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return persistentVolumeClaim, err
	}
}

func (p persistentVolumeClaimsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().PersistentVolumeClaim().Invalidate()
		return err
	}
}
