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

	pbImplSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2"
	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbSchedulerV2.Name, func() Integration {
		return &schedulerV2{}
	})
}

type schedulerV2 struct {
	Configuration pbImplSchedulerV2.Configuration
	Driver        string
}

func (b *schedulerV2) Name() string {
	return pbSchedulerV2.Name
}

func (b *schedulerV2) Description() string {
	return "SchedulerV2 Integration"
}

func (b *schedulerV2) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar(&b.Configuration.Namespace, "namespace", utilConstants.NamespaceWithDefault("default"), "Kubernetes Namespace"),
		fs.StringVar(&b.Configuration.Deployment, "deployment", "", "ArangoDeployment Name"),
		fs.StringVar(&b.Driver, "driver", string(helm.ConfigurationDriverSecret), "Helm Driver"),
	)
}

func (b *schedulerV2) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return nil, errors.Errorf("Unable to create Kubernetes Client")
	}

	helm, err := helm.NewClient(helm.Configuration{
		Namespace: b.Configuration.Namespace,
		Config:    client.Config(),
		Driver:    (*helm.ConfigurationDriver)(util.NewType(b.Driver)),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create Helm Client")
	}

	return pbImplSchedulerV2.New(client, helm, b.Configuration)
}

func (*schedulerV2) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
