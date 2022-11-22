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

package anonymous

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Impl interface {
	Anonymous(gvk schema.GroupVersionKind) (Interface, bool)
}

type Interface interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (meta.Object, error)

	Create(ctx context.Context, obj meta.Object, opts meta.CreateOptions) (meta.Object, error)
	Update(ctx context.Context, obj meta.Object, opts meta.UpdateOptions) (meta.Object, error)
	UpdateStatus(ctx context.Context, obj meta.Object, opts meta.UpdateOptions) (meta.Object, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result meta.Object, err error)
	Delete(ctx context.Context, name string, opts meta.DeleteOptions) error
}
