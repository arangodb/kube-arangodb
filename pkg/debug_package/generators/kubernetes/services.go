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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Services() shared.Factory {
	return shared.NewFactory("kubernetes-services", true, services)
}

func listServices(client kubernetes.Interface) func() ([]*core.Service, error) {
	return func() ([]*core.Service, error) {
		return ListObjects[*core.ServiceList, *core.Service](context.Background(), client.CoreV1().Services(cli.GetInput().Namespace), func(result *core.ServiceList) []*core.Service {
			q := make([]*core.Service, len(result.Items))

			for id, e := range result.Items {
				q[id] = e.DeepCopy()
			}

			return q
		})
	}
}

func services(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	services, err := listServices(k.Kubernetes())()
	if err != nil {
		return err
	}

	files <- shared.NewYAMLFile("kubernetes/services.yaml", func() ([]*core.Service, error) {
		return services, nil
	})

	for _, svc := range services {
		endpoints(k, svc.GetNamespace(), svc.GetName(), files)
	}

	return nil
}

func endpoints(k kclient.Client, namespace, name string, files chan<- shared.File) {
	ep, err := k.Kubernetes().CoreV1().Endpoints(namespace).Get(context.Background(), name, meta.GetOptions{})
	if err == nil {
		files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/services/%s/endpoints.yaml", name), func() ([]interface{}, error) {
			ep.ManagedFields = nil

			return []interface{}{ep}, nil
		})
	}
}
