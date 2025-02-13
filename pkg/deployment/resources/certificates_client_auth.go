//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
)

// createClientAuthCACertificate creates a client authentication CA certificate and stores it in a secret with name
// specified in the given spec.
func (r *Resources) createClientAuthCACertificate(ctx context.Context, secrets generic.ModClient[*core.Secret], spec api.SyncAuthenticationSpec, deploymentName string, ownerRef *meta.OwnerReference) error {
	log := r.log.Str("section", "secrets")

	cert, priv, err := ktls.CreateTLSCACertificate(fmt.Sprintf("%s Client Authentication Root Certificate", deploymentName))
	if err != nil {
		log.Err(err).Str("name", spec.GetClientCASecretName()).Debug("Failed to create CA certificate")
		return errors.WithStack(err)
	}
	if err := k8sutil.CreateCASecret(ctx, secrets, spec.GetClientCASecretName(), cert, priv, ownerRef); err != nil {
		if kerrors.IsAlreadyExists(err) {
			log.Debug("CA Secret already exists")
		} else {
			log.Err(err).Str("name", spec.GetClientCASecretName()).Debug("Failed to create CA Secret")
		}
		return errors.WithStack(err)
	}
	log.Str("name", spec.GetClientCASecretName()).Debug("Created CA Secret")
	return nil
}
