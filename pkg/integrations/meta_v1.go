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

	"github.com/spf13/cobra"

	pbImplMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	integrationsShared "github.com/arangodb/kube-arangodb/pkg/integrations/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbMetaV1.Name, func() Integration {
		return &metaV1{}
	})
}

type metaV1 struct {
	config pbImplMetaV1.Configuration
}

func (a metaV1) Name() string {
	return pbMetaV1.Name
}

func (a *metaV1) Description() string {
	return "Enable MetaV1 Integration Service"
}

func (a *metaV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar(&a.config.Prefix, "prefix", "", "Meta Key Prefix"),
		fs.DurationVar(&a.config.TTL, "ttl", 0, "Cache Object TTL"),
	)
}

func (a *metaV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	if err := integrationsShared.FillAll(cmd, &a.config.Endpoint, &a.config.Database); err != nil {
		return nil, err
	}

	return pbImplMetaV1.New(ctx, a.config)
}
