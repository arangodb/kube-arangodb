//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"
	"reflect"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type IterateOverItemsFunc[T meta.Object] func(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- File, item T) error

type Extract[T meta.Object] func(ctx context.Context, client kclient.Client, namespace string) ([]T, error)

func WithKubernetesItems[T meta.Object](extract Extract[T], iterators ...IterateOverItemsFunc[T]) GenFunc {
	return func(logger zerolog.Logger, files chan<- File) error {
		files, c := WithPrefix(files, "kubernetes/")
		defer c()

		k, ok := kclient.GetDefaultFactory().Client()
		if !ok {
			return errors.Errorf("Client is not initialised")
		}

		items, err := extract(shutdown.Context(), k, cli.GetInput().Namespace)
		if err != nil {
			if kerrors.IsForbiddenOrNotFound(err) {
				logger.Err(err).Msgf("Unable to list resources")
				return nil
			}

			return err
		}

		files, c = WithGVRPrefix(files, reflect.TypeOf(items).Elem())
		defer c()

		files <- NewYAMLFile(".yaml", func() ([]T, error) {
			return items, nil
		})

		for _, item := range items {
			cp, ok := k8sutil.Copy(item)
			if !ok {
				return errors.Errorf("Unable to copy item")
			}
			if err := WithItem[T](shutdown.Context(), logger, k, files, cp, iterators...); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithItem[T meta.Object](ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- File, item T, iterators ...IterateOverItemsFunc[T]) error {
	files, c := WithPrefix(files, "/%s/", item.GetName())
	defer c()
	for _, iter := range iterators {
		cp, ok := k8sutil.Copy(item)
		if !ok {
			return errors.Errorf("Unable to copy item")
		}
		if err := iter(ctx, logger, client, files, cp); err != nil {
			return err
		}
	}
	return nil
}

func WithModification[T meta.Object](in IterateOverItemsFunc[T], mods ...util.ModR[T]) IterateOverItemsFunc[T] {
	return func(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- File, item T) error {
		n := util.ApplyModsR(item, mods...)

		return in(ctx, logger, client, files, n)
	}
}

func WithDefinitions[T meta.Object](ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- File, item T) error {
	files <- NewYAMLFile("definition.yaml", func() ([]T, error) {
		return []T{item}, nil
	})

	return nil
}
