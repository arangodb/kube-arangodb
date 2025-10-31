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

package inventory

import (
	"context"

	"github.com/arangodb/go-driver"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
)

type Items []*Item

func FetchInventory(ctx context.Context, logger logging.Logger, threads int, conn driver.Connection, cfg *Configuration) (Items, error) {
	var out []*Item
	done := make(chan struct{})
	in := make(chan *Item)

	go func() {
		defer close(done)

		for z := range in {
			if z == nil {
				continue
			}

			out = append(out, z)
		}
	}()

	if err := executor.Run(ctx, logger, threads, runExecution(conn, cfg, in)); err != nil {
		return nil, err
	}

	close(in)

	<-done

	return out, nil
}

func runExecution(conn driver.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for _, executor := range global.Items() {
			log.Str("name", executor.K).Info("Starting executor")
			q := executor.V(conn, cfg, out)

			h.RunAsync(ctx, q)
		}

		h.WaitForSubThreads(t)

		return nil
	}
}

type Executor func(conn driver.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc

func (i *Item) Validate() error {
	if i == nil {
		return errors.Errorf("Item is not provided")
	}

	return errors.Errors(
		shared.ValidatePath("type", i.Type, func(s string) error {
			if s == "" {
				return errors.Errorf("Type cannot be empty")
			}

			return nil
		}),
		shared.ValidateRequiredInterfacePath("value", i.Value),
	)
}

func (i *ItemValue) Validate() error {
	if i == nil {
		return errors.Errorf("Item Value is not provided")
	}

	if i.GetValue() == nil {
		return errors.Errorf("Item Value is not provided")
	}

	return nil
}
