//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	jg "github.com/golang-jwt/jwt"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
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
func (r *Resources) EnsureSecrets(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	start := time.Now()
	spec := r.context.GetSpec()
	secrets := cachedStatus.SecretsModInterface().V1()
	status := r.context.GetStatus()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	image := status.CurrentImage
	imageFound := status.CurrentImage != nil
	defer metrics.SetDuration(inspectSecretsDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedSecretsCounters.WithLabelValues(deploymentName)

	log := r.log.Str("section", "secret")

	members := status.Members.AsList()

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	if spec.IsAuthenticated() {
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureTokenSecret(ctx, cachedStatus, secrets, spec.Authentication.GetJWTSecretName())); err != nil {
			return errors.Section(err, "JWT Secret")
		}
	}
	if spec.IsSecure() {
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureTLSCACertificateSecret(ctx, cachedStatus, secrets, spec.TLS)); err != nil {
			return errors.Section(err, "TLS CA")
		}
	}

	if err := reconcileRequired.Reconcile(ctx); err != nil {
		return err
	}

	if spec.IsAuthenticated() {
		if imageFound {
			if pod.VersionHasJWTSecretKeyfolder(image.ArangoDBVersion, image.Enterprise) {
				if err := r.ensureTokenSecretFolder(ctx, cachedStatus, secrets, spec.Authentication.GetJWTSecretName(), pod.JWTSecretFolder(deploymentName)); err != nil {
					return errors.Section(err, "JWT Folder")
				}
			}
		}

		if spec.Metrics.IsEnabled() {
			if imageFound && pod.VersionHasJWTSecretKeyfolder(image.ArangoDBVersion, image.Enterprise) {
				if err := reconcileRequired.WithError(r.ensureExporterTokenSecret(ctx, cachedStatus, secrets, spec.Metrics.GetJWTTokenSecretName(), pod.JWTSecretFolder(deploymentName))); err != nil {
					return errors.Section(err, "Metrics JWT")
				}
			} else {
				if err := reconcileRequired.WithError(r.ensureExporterTokenSecret(ctx, cachedStatus, secrets, spec.Metrics.GetJWTTokenSecretName(), spec.Authentication.GetJWTSecretName())); err != nil {
					return errors.Section(err, "Metrics JWT")
				}
			}
		}
	}
	if spec.IsSecure() {
		if err := reconcileRequired.WithError(r.ensureSecretWithEmptyKey(ctx, cachedStatus, secrets, GetCASecretName(r.context.GetAPIObject()), "empty")); err != nil {
			return errors.Section(err, "TLS TrustStore")
		}

		if err := reconcileRequired.ParallelAll(len(members), func(id int) error {
			if !members[id].Group.IsArangod() {
				return nil
			}

			memberName := members[id].Member.ArangoMemberName(r.context.GetAPIObject().GetName(), members[id].Group)

			member, ok := cachedStatus.ArangoMember().V1().GetSimple(memberName)
			if !ok {
				return errors.Newf("Member %s not found", memberName)
			}

			service, ok := cachedStatus.Service().V1().GetSimple(memberName)
			if !ok {
				return errors.Newf("Service of member %s not found", memberName)
			}

			tlsKeyfileSecretName := k8sutil.AppendTLSKeyfileSecretPostfix(member.GetName())
			if _, exists := cachedStatus.Secret().V1().GetSimple(tlsKeyfileSecretName); !exists {
				serverNames, err := tls.GetServerAltNames(apiObject, spec, spec.TLS, service, members[id].Group, members[id].Member)
				if err != nil {
					return errors.WithStack(errors.Wrapf(err, "Failed to render alt names"))
				}
				owner := member.AsOwner()
				if created, err := createTLSServerCertificate(ctx, log, cachedStatus, secrets, serverNames, spec.TLS, tlsKeyfileSecretName, &owner); err != nil && !kerrors.IsAlreadyExists(err) {
					return errors.WithStack(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
				} else if created {
					reconcileRequired.Required()
				}
			}
			return nil
		}); err != nil {
			return errors.Section(err, "TLS TrustStore")
		}
	}
	if spec.RocksDB.IsEncrypted() {
		if i := status.CurrentImage; i != nil && features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
			if err := reconcileRequired.WithError(r.ensureEncryptionKeyfolderSecret(ctx, cachedStatus, secrets, spec.RocksDB.Encryption.GetKeySecretName(), pod.GetEncryptionFolderSecretName(deploymentName))); err != nil {
				return errors.Section(err, "Encryption")
			}
		}
	}
	if r.context.IsSyncEnabled() {
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureTokenSecret(ctx, cachedStatus, secrets, spec.Sync.Authentication.GetJWTSecretName())); err != nil {
			return errors.Section(err, "Sync Auth")
		}
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureTokenSecret(ctx, cachedStatus, secrets, spec.Sync.Monitoring.GetTokenSecretName())); err != nil {
			return errors.Section(err, "Sync Monitoring Auth")
		}
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureTLSCACertificateSecret(ctx, cachedStatus, secrets, spec.Sync.TLS)); err != nil {
			return errors.Section(err, "Sync TLS CA")
		}
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureClientAuthCACertificateSecret(ctx, cachedStatus, secrets, spec.Sync.Authentication)); err != nil {
			return errors.Section(err, "Sync TLS Client CA")
		}
	}
	return reconcileRequired.Reconcile(ctx)
}

func (r *Resources) ensureTokenSecretFolder(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, secretName, folderSecretName string) error {
	if f, exists := cachedStatus.Secret().V1().GetSimple(folderSecretName); exists {
		if len(f.Data) == 0 {
			s, exists := cachedStatus.Secret().V1().GetSimple(secretName)
			if !exists {
				return errors.Newf("Token secret does not exist")
			}

			token, ok := s.Data[constants.SecretKeyToken]
			if !ok {
				return errors.Newf("Token secret is invalid")
			}

			f.Data[util.SHA256(token)] = token
			f.Data[pod.ActiveJWTKey] = token
			f.Data[constants.SecretKeyToken] = token

			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := secrets.Update(ctxChild, f, meta.UpdateOptions{})
				return err
			})
			if err != nil {
				return err
			}

			return errors.Reconcile()
		}

		if _, ok := f.Data[pod.ActiveJWTKey]; !ok {
			_, b, ok := getFirstKeyFromMap(f.Data)
			if !ok {
				return errors.Newf("Token Folder secret is invalid")
			}

			p := patch.NewPatch()
			p.ItemAdd(patch.NewPath("data", pod.ActiveJWTKey), util.SHA256(b))

			pdata, err := json.Marshal(p)
			if err != nil {
				return err
			}

			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := secrets.Patch(ctxChild, folderSecretName, types.JSONPatchType, pdata, meta.PatchOptions{})
				return err
			})
			if err != nil {
				return err
			}
		}

		if _, ok := f.Data[constants.SecretKeyToken]; !ok {
			b, ok := f.Data[pod.ActiveJWTKey]
			if !ok {
				return errors.Newf("Token Folder secret is invalid")
			}

			p := patch.NewPatch()
			p.ItemAdd(patch.NewPath("data", constants.SecretKeyToken), util.SHA256(b))

			pdata, err := json.Marshal(p)
			if err != nil {
				return err
			}

			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := secrets.Patch(ctxChild, folderSecretName, types.JSONPatchType, pdata, meta.PatchOptions{})
				return err
			})
			if err != nil {
				return err
			}
		}

		return nil
	}

	s, exists := cachedStatus.Secret().V1().GetSimple(secretName)
	if !exists {
		return errors.Newf("Token secret does not exist")
	}

	token, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		return errors.Newf("Token secret is invalid")
	}

	if err := r.createSecretWithMod(ctx, secrets, folderSecretName, func(s *core.Secret) {
		s.Data[util.SHA256(token)] = token
		s.Data[pod.ActiveJWTKey] = token
		s.Data[constants.SecretKeyToken] = token
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureTokenSecret(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, secretName string) error {
	if _, exists := cachedStatus.Secret().V1().GetSimple(secretName); !exists {
		return r.createTokenSecret(ctx, secrets, secretName)
	}

	return nil
}

func (r *Resources) ensureSecretWithEmptyKey(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, secretName, keyName string) error {
	if _, exists := cachedStatus.Secret().V1().GetSimple(secretName); !exists {
		return r.createSecretWithKey(ctx, secrets, secretName, keyName, nil)
	}

	return nil
}

func (r *Resources) createSecretWithMod(ctx context.Context, secrets secretv1.ModInterface, secretName string, f func(s *core.Secret)) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{},
	}
	// Attach secret to owner
	owner := r.context.GetAPIObject().AsOwner()
	k8sutil.AddOwnerRefToObject(secret, &owner)

	f(secret)

	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := secrets.Create(ctxChild, secret, meta.CreateOptions{})
		return kerrors.NewResourceError(err, secret)
	})
	if err != nil {
		// Failed to create secret
		return errors.WithStack(err)
	}

	return errors.Reconcile()
}

func (r *Resources) createSecretWithKey(ctx context.Context, secrets secretv1.ModInterface, secretName, keyName string, value []byte) error {
	return r.createSecretWithMod(ctx, secrets, secretName, func(s *core.Secret) {
		s.Data[keyName] = value
	})
}

func (r *Resources) createTokenSecret(ctx context.Context, secrets secretv1.ModInterface, secretName string) error {
	tokenData := make([]byte, 32)
	rand.Read(tokenData)
	token := hex.EncodeToString(tokenData)

	// Create secret
	owner := r.context.GetAPIObject().AsOwner()
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.CreateTokenSecret(ctxChild, secrets, secretName, token, &owner)
	})
	if kerrors.IsAlreadyExists(err) {
		// Secret added while we tried it also
		return nil
	} else if err != nil {
		// Failed to create secret
		return errors.WithStack(err)
	}

	return errors.Reconcile()
}

func (r *Resources) ensureEncryptionKeyfolderSecret(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, keyfileSecretName, secretName string) error {
	_, folderExists := cachedStatus.Secret().V1().GetSimple(secretName)

	keyfile, exists := cachedStatus.Secret().V1().GetSimple(keyfileSecretName)
	if !exists {
		if folderExists {
			return nil
		}
		return errors.Newf("Unable to find original secret %s", keyfileSecretName)
	}

	if len(keyfile.Data) == 0 {
		if folderExists {
			return nil
		}
		return errors.Newf("Missing key in secret")
	}

	d, ok := keyfile.Data[constants.SecretEncryptionKey]
	if !ok {
		if folderExists {
			return nil
		}
		return errors.Newf("Missing key in secret")
	}

	owner := r.context.GetAPIObject().AsOwner()
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return AppendKeyfileToKeyfolder(ctxChild, cachedStatus, secrets, &owner, secretName, d)
	})
	if err != nil {
		return errors.Wrapf(err, "Unable to create keyfolder secret")
	}
	return nil
}

func AppendKeyfileToKeyfolder(ctx context.Context, cachedStatus inspectorInterface.Inspector,
	secrets secretv1.ModInterface, ownerRef *meta.OwnerReference, secretName string, encryptionKey []byte) error {
	encSha := fmt.Sprintf("%0x", sha256.Sum256(encryptionKey))
	if _, exists := cachedStatus.Secret().V1().GetSimple(secretName); !exists {

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
		if _, err := secrets.Create(ctx, secret, meta.CreateOptions{}); err != nil {
			// Failed to create secret
			return kerrors.NewResourceError(err, secret)
		}

		return errors.Reconcile()
	}

	return nil
}

var (
	exporterTokenClaims = jg.MapClaims{
		"iss":       "arangodb",
		"server_id": "exporter",
		"allowed_paths": []interface{}{"/_admin/statistics", "/_admin/statistics-description",
			shared.ArangoExporterInternalEndpoint, shared.ArangoExporterInternalEndpointV2,
			shared.ArangoExporterStatusEndpoint, shared.ArangoExporterClusterHealthEndpoint},
	}
)

// ensureExporterTokenSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with correct access.
func (r *Resources) ensureExporterTokenSecret(ctx context.Context, cachedStatus inspectorInterface.Inspector,
	secrets secretv1.ModInterface, tokenSecretName, secretSecretName string) error {
	if update, exists, err := r.ensureExporterTokenSecretCreateRequired(cachedStatus, tokenSecretName, secretSecretName); err != nil {
		return err
	} else if update {
		// Create secret
		if !exists {
			owner := r.context.GetAPIObject().AsOwner()
			err = k8sutil.CreateJWTFromSecret(ctx, cachedStatus.Secret().V1().Read(), secrets, tokenSecretName, secretSecretName, exporterTokenClaims, &owner)
			if kerrors.IsAlreadyExists(err) {
				// Secret added while we tried it also
				return nil
			} else if err != nil {
				// Failed to create secret
				return errors.WithStack(err)
			}
		}

		return errors.Reconcile()
	}
	return nil
}

func (r *Resources) ensureExporterTokenSecretCreateRequired(cachedStatus inspectorInterface.Inspector, tokenSecretName, secretSecretName string) (bool, bool, error) {
	if secret, exists := cachedStatus.Secret().V1().GetSimple(tokenSecretName); !exists {
		return true, false, nil
	} else {
		// Check if claims are fine
		data, ok := secret.Data[constants.SecretKeyToken]
		if !ok {
			return true, true, nil
		}

		jwtSecret, exists := cachedStatus.Secret().V1().GetSimple(secretSecretName)
		if !exists {
			return true, true, errors.Newf("Secret %s does not exists", secretSecretName)
		}

		secret, err := k8sutil.GetTokenFromSecret(jwtSecret)
		if err != nil {
			return true, true, errors.WithStack(err)
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
func (r *Resources) ensureTLSCACertificateSecret(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, spec api.TLSSpec) error {
	if _, exists := cachedStatus.Secret().V1().GetSimple(spec.GetCASecretName()); !exists {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.createTLSCACertificate(ctxChild, secrets, spec, deploymentName, &owner)
		})
		if kerrors.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return errors.WithStack(err)
		}

		return errors.Reconcile()
	}
	return nil
}

// ensureClientAuthCACertificateSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a generated CA certificate.
func (r *Resources) ensureClientAuthCACertificateSecret(ctx context.Context, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, spec api.SyncAuthenticationSpec) error {
	if _, exists := cachedStatus.Secret().V1().GetSimple(spec.GetClientCASecretName()); !exists {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.createClientAuthCACertificate(ctxChild, secrets, spec, deploymentName, &owner)
		})
		if kerrors.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return errors.WithStack(err)
		}

		return errors.Reconcile()
	}
	return nil
}

// getJWTSecret loads the JWT secret from a Secret configured in apiObject.Spec.Authentication.JWTSecretName.
func (r *Resources) getJWTSecret(spec api.DeploymentSpec) (string, error) {
	if !spec.IsAuthenticated() {
		return "", nil
	}
	secretName := spec.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(context.Background(), r.context.ACS().CurrentClusterCache().Secret().V1().Read(), secretName)
	if err != nil {
		r.log.Str("section", "jwt").Err(err).Str("secret-name", secretName).Debug("Failed to get JWT secret")
		return "", errors.WithStack(err)
	}
	return s, nil
}

// getSyncJWTSecret loads the JWT secret used for syncmasters from a Secret configured in apiObject.Spec.Sync.Authentication.JWTSecretName.
func (r *Resources) getSyncJWTSecret(spec api.DeploymentSpec) (string, error) {
	secretName := spec.Sync.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(context.Background(), r.context.ACS().CurrentClusterCache().Secret().V1().Read(), secretName)
	if err != nil {
		r.log.Str("section", "jwt").Err(err).Str("secret-name", secretName).Debug("Failed to get sync JWT secret")
		return "", errors.WithStack(err)
	}
	return s, nil
}

// getSyncMonitoringToken loads the token secret used for monitoring sync masters & workers.
func (r *Resources) getSyncMonitoringToken(spec api.DeploymentSpec) (string, error) {
	secretName := spec.Sync.Monitoring.GetTokenSecretName()
	s, err := k8sutil.GetTokenSecret(context.Background(), r.context.ACS().CurrentClusterCache().Secret().V1().Read(), secretName)
	if err != nil {
		r.log.Str("section", "jwt").Err(err).Str("secret-name", secretName).Debug("Failed to get sync monitoring secret")
		return "", errors.WithStack(err)
	}
	return s, nil
}

func getFirstKeyFromMap(m map[string][]byte) (string, []byte, bool) {
	for k, v := range m {
		return k, v, true
	}

	return "", nil, false
}
