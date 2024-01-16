//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package svc

import (
	"context"
	"sync"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
)

type Service interface {
	Run(ctx context.Context) error
}

func RunServices(ctx context.Context, healthService probe.HealthService, services ...Service) error {
	if len(services) == 0 {
		<-ctx.Done()
		return nil
	}

	errors := make([]error, len(services))

	var wg sync.WaitGroup

	for id := range services {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			errors[id] = services[id].Run(ctx)

			healthService.Shutdown()
		}(id)
	}

	healthService.SetServing()
	wg.Wait()

	return shared.WithErrors(errors...)
}
