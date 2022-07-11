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

package v1

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/anonymous"
)

// ModInterface has methods to work with ServiceAccount resources only for creation
type ModInterface interface {
	Create(ctx context.Context, serviceaccount *core.ServiceAccount, opts meta.CreateOptions) (*core.ServiceAccount, error)
	Update(ctx context.Context, serviceaccount *core.ServiceAccount, opts meta.UpdateOptions) (*core.ServiceAccount, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *core.ServiceAccount, err error)
	Delete(ctx context.Context, name string, opts meta.DeleteOptions) error
}

// Interface has methods to work with ServiceAccount resources.
type Interface interface {
	anonymous.Impl

	ModInterface
	ReadInterface
}

// ReadInterface has methods to work with ServiceAccount resources with ReadOnly mode.
type ReadInterface interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (*core.ServiceAccount, error)
}
