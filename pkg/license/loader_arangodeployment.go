package license

import (
	"context"
	"encoding/base64"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewArengoDeploymentLicenseLoader(factory kclient.Factory, namespace, name string) Loader {
	return arangoDeploymentLicenseLoader{
		factory:   factory,
		namespace: namespace,
		name:      name,
	}
}

type arangoDeploymentLicenseLoader struct {
	factory kclient.Factory

	namespace, name string
}

func (a arangoDeploymentLicenseLoader) Refresh(ctx context.Context) (string, bool, error) {
	client, ok := a.factory.Client()
	if !ok {
		return "", false, nil
	}

	deployment, err := client.Arango().DatabaseV1().ArangoDeployments(a.namespace).Get(ctx, a.name, v1.GetOptions{})
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

	secret, err := client.Kubernetes().CoreV1().Secrets(deployment.GetNamespace()).Get(ctx, spec.License.GetSecretName(), v1.GetOptions{})
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
