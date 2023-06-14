//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Secrets() shared.Factory {
	return shared.NewFactory("kubernetes-secrets", true, secrets)
}

func listSecrets(client kubernetes.Interface) func() ([]*core.Secret, error) {
	return func() ([]*core.Secret, error) {
		return ListObjects[*core.SecretList, *core.Secret](context.Background(), client.CoreV1().Secrets(cli.GetInput().Namespace), func(result *core.SecretList) []*core.Secret {
			q := make([]*core.Secret, len(result.Items))

			for id, e := range result.Items {
				q[id] = e.DeepCopy()
			}

			return q
		})
	}
}

func secrets(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	files <- shared.NewYAMLFile("kubernetes/secrets.yaml", listSecrets(k.Kubernetes()))

	return nil
}
