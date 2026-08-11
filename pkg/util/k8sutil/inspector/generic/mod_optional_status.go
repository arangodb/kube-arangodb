//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package generic

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// NewModOptionalStatusClient wraps client so that UpdateStatus/PatchStatus target the status
// subresource when subresources reports it as exposed, and fall back to the normal Update/Patch if the
// subresource is not exposed or the apiserver reports it as NotFound.
func NewModOptionalStatusClient[S meta.Object](subresources SubresourceInspector, client ModStatusClient[S]) ModOptionalStatusClient[S] {
	return modOptionalStatusClient[S]{
		ModStatusClient: client,
		subresources:    subresources,
	}
}

type modOptionalStatusClient[S meta.Object] struct {
	ModStatusClient[S]

	subresources SubresourceInspector
}

func (c modOptionalStatusClient[S]) UpdateStatus(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	if c.subresources != nil && c.subresources.SubresourceEnabled(SubresourceStatus) {
		result, err := c.ModStatusClient.UpdateStatus(ctx, obj, opts)
		if !kerrors.IsNotFound(err) {
			return result, err
		}
		// The status subresource is not actually served - fall back to a normal Update.
	}

	return c.ModStatusClient.Update(ctx, obj, opts)
}

func (c modOptionalStatusClient[S]) PatchStatus(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions) (S, error) {
	if c.subresources != nil && c.subresources.SubresourceEnabled(SubresourceStatus) {
		result, err := c.ModStatusClient.Patch(ctx, name, pt, data, opts, SubresourceStatus.String())
		if !kerrors.IsNotFound(err) {
			return result, err
		}
		// The status subresource is not actually served - fall back to a normal Patch.
	}

	return c.ModStatusClient.Patch(ctx, name, pt, data, opts)
}
