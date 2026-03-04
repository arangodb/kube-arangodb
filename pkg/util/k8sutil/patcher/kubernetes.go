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

package patcher

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

type PatchInterface[P1 meta.Object] interface {
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (P1, error)
}

func WithKubernetesPatch[P1 meta.Object](ctx context.Context, obj string, client PatchInterface[P1], p ...patch.Item) (P1, error) {
	if len(p) == 0 {
		return util.Default[P1](), nil
	}

	parser := patch.Patch(p)

	data, err := parser.Marshal()
	if err != nil {
		return util.Default[P1](), err
	}

	nCtx, c := context.WithTimeout(ctx, globals.GetGlobals().Timeouts().Kubernetes().Get())
	defer c()

	return client.Patch(nCtx, obj, types.JSONPatchType, data, meta.PatchOptions{})
}
