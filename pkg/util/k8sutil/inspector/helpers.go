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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/refresh"
)

type ArangoMemberUpdateInterface interface {
	refresh.Inspector

	arangomember.Inspector
	ArangoMemberModInterface() mods.ArangoMemberMods
}

type ArangoMemberUpdateFunc func(in *api.ArangoMember) (bool, error)

func WithArangoMemberUpdate(ctx context.Context, client ArangoMemberUpdateInterface, name string, f ArangoMemberUpdateFunc) error {
	nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	obj, err := client.ArangoMember().V1().Read().Get(nctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	nobj := obj.DeepCopy()

	if changed, err := f(nobj); err != nil {
		return err
	} else if changed {
		nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
		defer c()

		if _, err := client.ArangoMemberModInterface().V1().Update(nctx, nobj, meta.UpdateOptions{}); err != nil {
			return err
		}

		return client.Refresh(ctx)
	}

	return nil
}

func WithArangoMemberStatusUpdate(ctx context.Context, client ArangoMemberUpdateInterface, name string, f ArangoMemberUpdateFunc) error {
	nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	obj, err := client.ArangoMember().V1().Read().Get(nctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	nobj := obj.DeepCopy()

	if changed, err := f(nobj); err != nil {
		return err
	} else if changed {
		nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
		defer c()

		if _, err := client.ArangoMemberModInterface().V1().UpdateStatus(nctx, nobj, meta.UpdateOptions{}); err != nil {
			return err
		}

		return client.Refresh(ctx)
	}

	return nil
}
