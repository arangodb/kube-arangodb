//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	pbImplConfigV1 "github.com/arangodb/kube-arangodb/integrations/config/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplConfigV1.Name, func() Integration {
		return &configV1{}
	})
}

type configV1 struct {
	modules []string
}

func (a *configV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringSliceVar(&a.modules, "module", nil, "Module in the reference <name>=<abs path>"),
	)
}

func (a *configV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	var cfg pbImplConfigV1.Config

	cfg.Modules = map[string]pbImplConfigV1.ModuleDefinition{}

	for _, module := range a.modules {
		l := strings.SplitN(module, "=", 2)
		if len(l) != 2 {
			return nil, errors.Errorf("Invalid module definition: %s", module)
		}

		cfg.Modules[l[0]] = pbImplConfigV1.ModuleDefinition{
			Path: l[1],
		}
	}

	return pbImplConfigV1.New(cfg)
}

func (a *configV1) Name() string {
	return pbImplConfigV1.Name
}

func (a *configV1) Description() string {
	return "Enable ConfigV1 Integration Service"
}

func (*configV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
