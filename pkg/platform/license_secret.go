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

package platform

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	lmanager "github.com/arangodb/kube-arangodb/pkg/license_manager"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func licenseSecret() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "secret"
	cmd.Short = "Creates Platform Secret with Registry credentials"

	if err := cli.RegisterFlags(&cmd, flagSecret, flagLicenseManager); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(licenseSecretRun).Run

	return &cmd, nil
}

func licenseSecretRun(cmd *cobra.Command, args []string) error {
	client, err := getKubernetesClient(cmd)
	if err != nil {
		return err
	}

	name, err := flagSecret.Get(cmd)
	if err != nil {
		return err
	}

	namespace, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	stages, err := flagLicenseManager.Stages(cmd)
	if err != nil {
		return err
	}

	id, err := flagLicenseManager.ClientID(cmd)
	if err != nil {
		return err
	}

	endpoint, err := flagLicenseManager.Endpoint(cmd)
	if err != nil {
		return err
	}

	mc, err := flagLicenseManager.Client(cmd)
	if err != nil {
		return err
	}

	logger.Info("Creating new Registry Token")

	data, err := mc.RegistryConfig(cmd.Context(), endpoint, id, nil, lmanager.ParseStages(stages...)...)
	if err != nil {
		return err
	}

	if name != "" {
		sClient := client.Kubernetes().CoreV1().Secrets(namespace)

		l := logger.Str("namespace", namespace).Str("secret", name)

		if s, err := sClient.Get(cmd.Context(), name, meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}

			l.Info("Secret not found, creating")

			if _, err := sClient.Create(cmd.Context(), &core.Secret{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Type: core.SecretTypeDockerConfigJson,
				Data: map[string][]byte{
					".dockerconfigjson": data,
				},
			}, meta.CreateOptions{}); err != nil {
				return err
			}

			l.Info("Secret Created")
		} else {
			l.Info("Secret found, updating")

			s.Data = map[string][]byte{
				".dockerconfigjson": data,
			}

			if _, err := sClient.Update(cmd.Context(), s, meta.UpdateOptions{}); err != nil {
				return err
			}

			l.Info("Secret Updated")
		}
	} else {
		resp, err := yaml.Marshal(&core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name: "name",
			},
			Type: core.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				".dockerconfigjson": data,
			},
		})
		if err != nil {
			return err
		}

		logger.Info("Create Secret Manually. Secret printed to STDERR")

		fmt.Fprint(os.Stderr, string(resp))
	}

	return nil
}
