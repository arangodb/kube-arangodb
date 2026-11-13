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

package sidecar

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	pbImplMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var (
	flagMetaV1Prefix = cli.Flag[string]{
		Name:        "meta.v1.prefix",
		Description: "Prefix applied to all meta keys",
		Default:     "",
	}
	flagMetaV1TTL = cli.Flag[time.Duration]{
		Name:        "meta.v1.ttl",
		Description: "Default TTL for meta entries (0 means no TTL)",
		Default:     0,
	}
)

func init() {
	register("meta-v1", func(ctx context.Context, cmd *cobra.Command) (svc.Handler, bool, error) {
		if v, err := flagCentralServicesEnabled.Get(cmd); err != nil {
			return nil, false, err
		} else if !v {
			return nil, false, nil
		}

		prefix, err := flagMetaV1Prefix.Get(cmd)
		if err != nil {
			return nil, false, err
		}

		ttl, err := flagMetaV1TTL.Get(cmd)
		if err != nil {
			return nil, false, err
		}

		svc, err := pbImplMetaV1.New(ctx, pbImplMetaV1.Configuration{
			Prefix: prefix,
			TTL:    ttl,
		})
		if err != nil {
			return nil, false, err
		}

		svc, err = pbImplMetaV1.WithRBAC(ctx, svc)
		if err != nil {
			return nil, false, err
		}

		return svc, true, nil
	},
		flagMetaV1Prefix,
		flagMetaV1TTL,
	)
}
