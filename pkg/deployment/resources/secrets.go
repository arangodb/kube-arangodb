//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package resources

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/rs/zerolog"

	operatorErrors "github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	jg "github.com/dgrijalva/jwt-go"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	inspectedSecretsCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_secrets", "Number of Secret inspections per deployment", metrics.DeploymentName)
	inspectSecretsDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_secrets_duration", "Amount of time taken by a single inspection of all Secrets for a deployment (in sec)", metrics.DeploymentName)
)

const (
	CAKeyName  = "ca.key"
	CACertName = "ca.crt"
)

func GetCASecretName(apiObject k8sutil.APIObject) string {
	return fmt.Sprintf("%s-truststore", apiObject.GetName())
}

// EnsureSecrets creates all secrets needed to run the given deployment
func (r *Resources) EnsureSecrets(log zerolog.Logger, cachedStatus inspector.Inspector) error {
	start := time.Now()
	spec := r.context.GetSpec()
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	status, _ := r.context.GetStatus()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	image := status.CurrentImage
	imageFound := status.CurrentImage != nil
	defer metrics.SetDuration(inspectSecretsDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedSecretsCounters.WithLabelValues(deploymentName)

	if spec.IsAuthenticated() {
		counterMetric.Inc()
		if err := r.ensureTokenSecret(cachedStatus, secrets, spec.Authentication.GetJWTSecretName()); err != nil {
			return maskAny(err)
		}

		if imageFound {
			if pod.VersionHasJWTSecretKeyfolder(image.ArangoDBVersion, image.Enterprise) {
				if err := r.ensureTokenSecretFolder(cachedStatus, secrets, spec.Authentication.GetJWTSecretName(), pod.JWTSecretFolder(deploymentName)); err != nil {
					return maskAny(err)
				}
			}
		}

		if spec.Metrics.IsEnabled() {
			if err := r.ensureExporterTokenSecret(cachedStatus, secrets, spec.Metrics.GetJWTTokenSecretName(), spec.Authentication.GetJWTSecretName()); err != nil {
				return maskAny(err)
			}
		}
	}
	if spec.IsSecure() {
		counterMetric.Inc()
		if err := r.ensureTLSCACertificateSecret(cachedStatus, secrets, spec.TLS); err != nil {
			return maskAny(err)
		}

		if err := r.ensureSecretWithEmptyKey(cachedStatus, secrets, GetCASecretName(r.context.GetAPIObject()), "empty"); err != nil {
			return maskAny(err)
		}

		if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
			if !group.IsArangod() {
				return nil
			}

			role := group.AsRole()

			for _, m := range list {
				tlsKeyfileSecretName := k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
				if _, exists := cachedStatus.Secret(tlsKeyfileSecretName); !exists {
					serverNames := []string{
						k8sutil.CreateDatabaseClientServiceDNSName(apiObject),
						k8sutil.CreatePodDNSName(apiObject, role, m.ID),
					}
					if ip := spec.ExternalAccess.GetLoadBalancerIP(); ip != "" {
						serverNames = append(serverNames, ip)
					}
					owner := apiObject.AsOwner()
					if err := createTLSServerCertificate(log, secrets, serverNames, spec.TLS, tlsKeyfileSecretName, &owner); err != nil && !k8sutil.IsAlreadyExists(err) {
						return maskAny(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
					}

					return operatorErrors.Reconcile()
				}
			}
			return nil
		}); err != nil {
			return maskAny(err)
		}
	}
	if spec.RocksDB.IsEncrypted() {
		if i := status.CurrentImage; i != nil && i.Enterprise && i.ArangoDBVersion.CompareTo("3.7.0") >= 0 {
			if err := r.ensureEncryptionKeyfolderSecret(cachedStatus, secrets, spec.RocksDB.Encryption.GetKeySecretName(), pod.GetEncryptionFolderSecretName(deploymentName)); err != nil {
				return maskAny(err)
			}
		}
	}
	if spec.Sync.IsEnabled() {
		counterMetric.Inc()
		if err := r.ensureTokenSecret(cachedStatus, secrets, spec.Sync.Authentication.GetJWTSecretName()); err != nil {
			return maskAny(err)
		}
		counterMetric.Inc()
		if err := r.ensureTokenSecret(cachedStatus, secrets, spec.Sync.Monitoring.GetTokenSecretName()); err != nil {
			return maskAny(err)
		}
		counterMetric.Inc()
		if err := r.ensureTLSCACertificateSecret(cachedStatus, secrets, spec.Sync.TLS); err != nil {
			return maskAny(err)
		}
		counterMetric.Inc()
		if err := r.ensureClientAuthCACertificateSecret(cachedStatus, secrets, spec.Sync.Authentication); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

func (r *Resources) ensureTokenSecretFolder(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, secretName, folderSecretName string) error {
	if _, exists := cachedStatus.Secret(folderSecretName); exists {
		return nil
	}

	s, exists := cachedStatus.Secret(secretName)
	if !exists {
		return errors.Errorf("Token secret does not exist")
	}

	token, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		return errors.Errorf("Token secret is invalid")
	}

	if err := r.createSecretWithKey(secrets, folderSecretName, util.SHA256(token), token); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureTokenSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, secretName string) error {
	if _, exists := cachedStatus.Secret(secretName); !exists {
		return r.createTokenSecret(secrets, secretName)
	}

	return nil
}

func (r *Resources) ensureSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, secretName string) error {
	if _, exists := cachedStatus.Secret(secretName); !exists {
		return r.createSecret(secrets, secretName)
	}

	return nil
}

func (r *Resources) createSecret(secrets k8sutil.SecretInterface, secretName string) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
	}
	// Attach secret to owner
	owner := r.context.GetAPIObject().AsOwner()
	k8sutil.AddOwnerRefToObject(secret, &owner)
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}

	return operatorErrors.Reconcile()
}

func (r *Resources) ensureSecretWithEmptyKey(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, secretName, keyName string) error {
	if _, exists := cachedStatus.Secret(secretName); !exists {
		return r.createSecretWithKey(secrets, secretName, keyName, nil)
	}

	return nil
}

func (r *Resources) ensureSecretWithKey(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, secretName, keyName string, value []byte) error {
	if _, exists := cachedStatus.Secret(secretName); !exists {
		return r.createSecretWithKey(secrets, secretName, keyName, value)
	}

	return nil
}

func (r *Resources) createSecretWithKey(secrets k8sutil.SecretInterface, secretName, keyName string, value []byte) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			keyName: value,
		},
	}
	// Attach secret to owner
	owner := r.context.GetAPIObject().AsOwner()
	k8sutil.AddOwnerRefToObject(secret, &owner)
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}

	return operatorErrors.Reconcile()
}

func (r *Resources) createTokenSecret(secrets k8sutil.SecretInterface, secretName string) error {
	tokenData := make([]byte, 32)
	rand.Read(tokenData)
	token := hex.EncodeToString(tokenData)

	// Create secret
	owner := r.context.GetAPIObject().AsOwner()
	if err := k8sutil.CreateTokenSecret(secrets, secretName, token, &owner); k8sutil.IsAlreadyExists(err) {
		// Secret added while we tried it also
		return nil
	} else if err != nil {
		// Failed to create secret
		return maskAny(err)
	}

	return operatorErrors.Reconcile()
}

func (r *Resources) ensureEncryptionKeyfolderSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, keyfileSecretName, secretName string) error {
	_, folderExists := cachedStatus.Secret(secretName)

	keyfile, exists := cachedStatus.Secret(keyfileSecretName)
	if !exists {
		if folderExists {
			return nil
		}
		return errors.Errorf("Unable to find original secret %s", keyfileSecretName)
	}

	if len(keyfile.Data) == 0 {
		if folderExists {
			return nil
		}
		return errors.Errorf("Missing key in secret")
	}

	d, ok := keyfile.Data[constants.SecretEncryptionKey]
	if !ok {
		if folderExists {
			return nil
		}
		return errors.Errorf("Missing key in secret")
	}

	owner := r.context.GetAPIObject().AsOwner()
	if err := AppendKeyfileToKeyfolder(cachedStatus, secrets, &owner, secretName, d); err != nil {
		return errors.Wrapf(err, "Unable to create keyfolder secret")
	}
	return nil
}

func AppendKeyfileToKeyfolder(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, ownerRef *meta.OwnerReference, secretName string, encryptionKey []byte) error {
	encSha := fmt.Sprintf("%0x", sha256.Sum256(encryptionKey))
	if _, exists := cachedStatus.Secret(secretName); !exists {

		// Create secret
		secret := &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				encSha: encryptionKey,
			},
		}
		// Attach secret to owner
		k8sutil.AddOwnerRefToObject(secret, ownerRef)
		if _, err := secrets.Create(secret); err != nil {
			// Failed to create secret
			return maskAny(err)
		}

		return operatorErrors.Reconcile()
	}

	return nil
}

var (
	exporterTokenClaims = jg.MapClaims{
		"iss":           "arangodb",
		"server_id":     "exporter",
		"allowed_paths": []interface{}{"/_admin/statistics", "/_admin/statistics-description", k8sutil.ArangoExporterInternalEndpoint},
	}
)

// ensureExporterTokenSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with correct access.
func (r *Resources) ensureExporterTokenSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, tokenSecretName, secretSecretName string) error {
	if recreate, exists, err := r.ensureExporterTokenSecretCreateRequired(cachedStatus, tokenSecretName, secretSecretName); err != nil {
		return err
	} else if recreate {
		// Create secret
		if exists {
			if err := secrets.Delete(tokenSecretName, nil); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}

		owner := r.context.GetAPIObject().AsOwner()
		if err := k8sutil.CreateJWTFromSecret(secrets, tokenSecretName, secretSecretName, exporterTokenClaims, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}

		return operatorErrors.Reconcile()
	}
	return nil
}

func (r *Resources) ensureExporterTokenSecretCreateRequired(cachedStatus inspector.Inspector, tokenSecretName, secretSecretName string) (bool, bool, error) {
	if secret, exists := cachedStatus.Secret(tokenSecretName); !exists {
		return true, false, nil
	} else {
		// Check if claims are fine
		data, ok := secret.Data[constants.SecretKeyToken]
		if !ok {
			return true, true, nil
		}

		jwtSecret, exists := cachedStatus.Secret(secretSecretName)
		if !exists {
			return true, true, errors.Errorf("Secret %s does not exists", secretSecretName)
		}

		secret, err := k8sutil.GetTokenFromSecret(jwtSecret)
		if err != nil {
			return true, true, maskAny(err)
		}

		token, err := jg.Parse(string(data), func(token *jg.Token) (i interface{}, err error) {
			return []byte(secret), nil
		})

		if err != nil {
			return true, true, nil
		}

		tokenClaims, ok := token.Claims.(jg.MapClaims)
		if !ok {
			return true, true, nil
		}

		return !equality.Semantic.DeepDerivative(tokenClaims, exporterTokenClaims), true, nil
	}
}

// ensureTLSCACertificateSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a generated CA certificate.
func (r *Resources) ensureTLSCACertificateSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, spec api.TLSSpec) error {
	if _, exists := cachedStatus.Secret(spec.GetCASecretName()); !exists {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		if err := createTLSCACertificate(r.log, secrets, spec, deploymentName, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}

		return operatorErrors.Reconcile()
	}
	return nil
}

// ensureClientAuthCACertificateSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a generated CA certificate.
func (r *Resources) ensureClientAuthCACertificateSecret(cachedStatus inspector.Inspector, secrets k8sutil.SecretInterface, spec api.SyncAuthenticationSpec) error {
	if _, exists := cachedStatus.Secret(spec.GetClientCASecretName()); !exists {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		if err := createClientAuthCACertificate(r.log, secrets, spec, deploymentName, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}

		return operatorErrors.Reconcile()
	}
	return nil
}

// getJWTSecret loads the JWT secret from a Secret configured in apiObject.Spec.Authentication.JWTSecretName.
func (r *Resources) getJWTSecret(spec api.DeploymentSpec) (string, error) {
	if !spec.IsAuthenticated() {
		return "", nil
	}
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := spec.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncJWTSecret loads the JWT secret used for syncmasters from a Secret configured in apiObject.Spec.Sync.Authentication.JWTSecretName.
func (r *Resources) getSyncJWTSecret(spec api.DeploymentSpec) (string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := spec.Sync.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncMonitoringToken loads the token secret used for monitoring sync masters & workers.
func (r *Resources) getSyncMonitoringToken(spec api.DeploymentSpec) (string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := spec.Sync.Monitoring.GetTokenSecretName()
	s, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync monitoring secret")
		return "", maskAny(err)
	}
	return s, nil
}
