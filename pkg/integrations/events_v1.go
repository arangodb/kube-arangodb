//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	pbImplEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1"
	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	integrationsShared "github.com/arangodb/kube-arangodb/pkg/integrations/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbEventsV1.Name, func() Integration {
		return &eventsV1{}
	})
}

type eventsV1 struct {
	config pbImplEventsV1.Configuration
}

func (a eventsV1) Name() string {
	return pbEventsV1.Name
}

func (a *eventsV1) Description() string {
	return "Enable EventsV1 Integration Service"
}

func (a *eventsV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.BoolVar(&a.config.Async.Enabled, "async", true, "Enables async injection of the events"),
		fs.IntVar(&a.config.Async.Size, "async.size", 16, "Size of the async queue"),
		fs.DurationVar(&a.config.Async.Retry.Delay, "async.retry.delay", time.Second, "Delay of the retries"),
		fs.DurationVar(&a.config.Async.Retry.Timeout, "async.retry.timeout", time.Minute, "Timeout for the event injection"),
	)
}

func (a *eventsV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	if err := integrationsShared.FillAll(cmd, &a.config.Endpoint, &a.config.Database); err != nil {
		return nil, err
	}

	return pbImplEventsV1.New(a.config)
}
