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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	typedApi "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1"
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
)

func (p arangoMemberMod) V1() arangomemberv1.ModInterface {
	return arangoMemberModV1(p)
}

type arangoMemberModV1 struct {
	i *inspectorState
}

func (p arangoMemberModV1) client() typedApi.ArangoMemberInterface {
	return p.i.Client().Arango().DatabaseV1().ArangoMembers(p.i.Namespace())
}

func (p arangoMemberModV1) Create(ctx context.Context, endpoint *api.ArangoMember, opts meta.CreateOptions) (*api.ArangoMember, error) {
	if endpoint, err := p.client().Create(ctx, endpoint, opts); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().ArangoMember().Invalidate()
		return endpoint, err
	}
}

func (p arangoMemberModV1) Update(ctx context.Context, endpoint *api.ArangoMember, opts meta.UpdateOptions) (*api.ArangoMember, error) {
	if endpoint, err := p.client().Update(ctx, endpoint, opts); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().ArangoMember().Invalidate()
		return endpoint, err
	}
}

func (p arangoMemberModV1) UpdateStatus(ctx context.Context, endpoint *api.ArangoMember, opts meta.UpdateOptions) (*api.ArangoMember, error) {
	if endpoint, err := p.client().UpdateStatus(ctx, endpoint, opts); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().ArangoMember().Invalidate()
		return endpoint, err
	}
}

func (p arangoMemberModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *api.ArangoMember, err error) {
	if endpoint, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return endpoint, err
	} else {
		p.i.GetThrottles().ArangoMember().Invalidate()
		return endpoint, err
	}
}

func (p arangoMemberModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().ArangoMember().Invalidate()
		return err
	}
}
