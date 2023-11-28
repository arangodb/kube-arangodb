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

package license

import (
	"context"
	"encoding/base64"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func NewArangoDeploymentLicenseLoader(client arangoClientSet.Interface, kubeClient kubernetes.Interface, namespace, name string) Loader {
	return arangoDeploymentLicenseLoader{
		client:     client,
		kubeClient: kubeClient,
		namespace:  namespace,
		name:       name,
	}
}

type arangoDeploymentLicenseLoader struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	namespace, name string
}

func (a arangoDeploymentLicenseLoader) Refresh(ctx context.Context) (string, bool, error) {
	deployment, err := a.client.DatabaseV1().ArangoDeployments(a.namespace).Get(ctx, a.name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", false, nil
		}

		return "", false, err
	}

	spec := deployment.GetAcceptedSpec()

	if !spec.License.HasSecretName() {
		return "", false, nil
	}

	secret, err := a.kubeClient.CoreV1().Secrets(deployment.GetNamespace()).Get(ctx, spec.License.GetSecretName(), meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", false, nil
		}

		return "", false, err
	}

	var licenseData []byte

	if lic, ok := secret.Data[constants.SecretKeyV2License]; ok {
		licenseData = lic
	} else if lic2, ok := secret.Data[constants.SecretKeyV2Token]; ok {
		licenseData = lic2
	}

	if len(licenseData) == 0 {
		return "", false, nil
	}

	if !k8sutil.IsJSON(licenseData) {
		d, err := base64.StdEncoding.DecodeString(string(licenseData))
		if err != nil {
			return "", false, err
		}

		licenseData = d
	}

	return string(licenseData), true, nil
}
