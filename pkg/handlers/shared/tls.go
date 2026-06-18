//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"
	"fmt"
	goStrings "strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb-helper/go-certificates"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helpers"
)

func OperatorTLSFromArangoDeployment(handler helpers.Updator[*core.Secret], ctx context.Context, client kubernetes.Interface, namespace string, parent k8sutil.K8SObject, ref **sharedApi.Object,
	arangoDeployment *api.ArangoDeployment, spec api.TLSSpec,
	tls *sharedApi.TLS) (bool, error) {
	if _, changed, err := handler.OperatorUpdate(ctx, namespace, parent, ref, func(ctx context.Context, ref *sharedApi.Object) (*core.Secret, bool, string, error) {
		if !(spec.IsSecure() && spec.GetCASecretName() != "") {
			// Ensure we remove secret
			return nil, false, "", nil
		}
		if !tls.IsEnabled() {
			// Ensure we remove secret if TLS is not enabled
			return nil, false, "", nil
		}
		ca, checksum, err := extractDeploymentCA(ctx, client, arangoDeployment)
		if err != nil {
			return nil, false, "", err
		}
		if ref != nil && ref.GetChecksum() == checksum {
			// Checksum is fine, no need for recreation
			return nil, true, "", nil
		}
		altNames := append([]string{
			parent.GetName(),
			fmt.Sprintf("%s.%s.svc", parent.GetName(), parent.GetNamespace()),
		}, tls.GetAltNames()...)
		options := certificates.CreateCertificateOptions{
			CommonName: altNames[0],
			Hosts:      altNames,
			ValidFrom:  time.Now(),
			ValidFor:   spec.GetTTL().AsDuration(),
			IsCA:       false,
			ECDSACurve: "P256",
		}
		cert, priv, err := certificates.CreateCertificate(options, &ca)
		if err != nil {
			return nil, false, "", err
		}
		keyfile := goStrings.TrimSpace(cert) + "\n" +
			goStrings.TrimSpace(priv)
		return k8sutil.RenderTLSKeyfileSecret(fmt.Sprintf("%s-tls", parent.GetName()), keyfile, util.NewType(parent.AsOwner())), false, checksum, nil
	}, helpers.UpdateOwnerReference[*core.Secret], helpers.ReplaceChecksum[*core.Secret]); err != nil {
		return changed, err
	}
	return false, nil
}
func extractDeploymentCA(ctx context.Context, client kubernetes.Interface, deployment *api.ArangoDeployment) (certificates.CA, string, error) {
	secret, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.CoreV1().Secrets(deployment.GetNamespace()).Get, deployment.GetAcceptedSpec().TLS.GetCASecretName(), meta.GetOptions{})
	if err != nil {
		return certificates.CA{}, "", err
	}
	caCert, caKey, _, err := k8sutil.GetCAFromSecret(secret, nil)
	if err != nil {
		return certificates.CA{}, "", err
	}
	ca, err := certificates.LoadCAFromPEM(caCert, caKey)
	if err != nil {
		return certificates.CA{}, "", err
	}
	return ca, util.MD5FromString(caCert), nil
}
