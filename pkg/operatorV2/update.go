//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

type Object[T interface{}] interface {
	meta.Object

	GetStatus() T
	SetStatus(T)
}

type GetInterface[S interface{}, T Object[S]] interface {
	Get(ctx context.Context, name string, options meta.GetOptions) (T, error)
}

type UpdateStatusInterfaceClient[S interface{}, T Object[S]] func(namespace string) UpdateStatusInterface[S, T]

type UpdateStatusInterface[S interface{}, T Object[S]] interface {
	GetInterface[S, T]

	UpdateStatus(ctx context.Context, in T, options meta.UpdateOptions) (T, error)
}

func WithUpdateStatusInterfaceRetry[S interface{}, T Object[S]](ctx context.Context, client UpdateStatusInterface[S, T], obj T, status S, opts meta.UpdateOptions) (T, error) {
	for id := 0; id < globals.GetGlobals().Retry().OperatorUpdateRetryCount().Get(); id++ {
		// Let's try to make a call
		if nObj, err := WithUpdateStatusInterface(ctx, client, obj, status, opts); err == nil {
			return nObj, nil
		}

		select {
		case <-timer.After(globals.GetGlobals().Retry().OperatorUpdateRetryDelay().Get()):
			continue
		case <-ctx.Done():
			return util.Default[T](), context.DeadlineExceeded
		}
	}

	return util.Default[T](), errors.Newf("Unable to save Object %s/%s, retries exceeded", obj.GetNamespace(), obj.GetName())
}

func WithUpdateStatusInterface[S interface{}, T Object[S]](ctx context.Context, client UpdateStatusInterface[S, T], obj T, status S, opts meta.UpdateOptions) (T, error) {
	cCtx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	currentObj, err := client.Get(cCtx, obj.GetName(), meta.GetOptions{})
	if err != nil {
		return util.Default[T](), err
	}

	currentObj.SetStatus(status)

	nCtx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	return client.UpdateStatus(nCtx, currentObj, opts)
}
