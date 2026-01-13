//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	authorization "k8s.io/api/authorization/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/access"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbSchedulerV1.SchedulerV1Server = &implementation{}
var _ svc.Handler = &implementation{}

func New(ctx context.Context, client kclient.Client, cfg Configuration) (svc.Handler, error) {
	return newInternal(ctx, client, cfg)
}

func newInternal(ctx context.Context, client kclient.Client, cfg Configuration) (*implementation, error) {
	if cfg.VerifyAccess {
		// Lets Verify Access
		if !access.VerifyAllAccess(ctx, client,
			authorization.ResourceAttributes{
				Verb:      "create",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerbatchjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "list",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerbatchjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "delete",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerbatchjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "get",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerbatchjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "create",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulercronjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "list",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulercronjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "delete",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulercronjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "get",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulercronjobs",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "create",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerpods",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "list",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerpods",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "delete",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerpods",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "get",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerpods",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "create",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerdeployments",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "list",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerdeployments",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "delete",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerdeployments",
				Namespace: cfg.Namespace,
			},
			authorization.ResourceAttributes{
				Verb:      "get",
				Group:     "scheduler.arangodb.com",
				Version:   "v1",
				Resource:  "arangoschedulerdeployments",
				Namespace: cfg.Namespace,
			},
		) {
			return nil, errors.Errorf("Unable to access API")
		}
	}

	return &implementation{
		cfg:    cfg,
		client: client,
	}, nil
}

type implementation struct {
	cfg Configuration

	client kclient.Client

	pbSchedulerV1.UnimplementedSchedulerV1Server
}

func (i *implementation) Name() string {
	return pbSchedulerV1.Name
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbSchedulerV1.RegisterSchedulerV1Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (i *implementation) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}
