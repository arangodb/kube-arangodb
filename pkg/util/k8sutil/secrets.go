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

package k8sutil

import (
	"context"

	jg "github.com/golang-jwt/jwt"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// ValidateEncryptionKeySecret checks that a secret with given name in given namespace
// exists and it contains a 'key' data field of exactly 32 bytes.
func ValidateEncryptionKeySecret(secrets secretv1.Interface, secretName string) error {
	s, err := secrets.Get(context.Background(), secretName, meta.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return ValidateEncryptionKeyFromSecret(s)
}

func ValidateEncryptionKeyFromSecret(s *core.Secret) error {
	// Check `key` field
	keyData, found := s.Data[constants.SecretEncryptionKey]
	if !found {
		return errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretEncryptionKey, s.GetName()))
	}
	if len(keyData) != 32 {
		return errors.WithStack(errors.Newf("'%s' in secret '%s' is expected to be 32 bytes long, found %d", constants.SecretEncryptionKey, s.GetName(), len(keyData)))
	}
	return nil
}

// CreateEncryptionKeySecret creates a secret used to store a RocksDB encryption key.
func CreateEncryptionKeySecret(secrets secretv1.ModInterface, secretName string, key []byte) error {
	if len(key) != 32 {
		return errors.WithStack(errors.Newf("Key in secret '%s' is expected to be 32 bytes long, got %d", secretName, len(key)))
	}
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretEncryptionKey: key,
		},
	}
	if _, err := secrets.Create(context.Background(), secret, meta.CreateOptions{}); err != nil {
		// Failed to create secret
		return kerrors.NewResourceError(err, secret)
	}
	return nil
}

// ValidateCACertificateSecret checks that a secret with given name in given namespace
// exists and it contains a 'ca.crt' data field.
func ValidateCACertificateSecret(ctx context.Context, secrets secretv1.ReadInterface, secretName string) error {
	s, err := secrets.Get(ctx, secretName, meta.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	// Check `ca.crt` field
	_, found := s.Data[constants.SecretCACertificate]
	if !found {
		return errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretCACertificate, secretName))
	}
	return nil
}

// GetCACertficateSecret loads a secret with given name in the given namespace
// and extracts the `ca.crt` field.
// If the secret does not exists the field is missing,
// an error is returned.
// Returns: certificate, error
func GetCACertficateSecret(ctx context.Context, secrets secretv1.ReadInterface, secretName string) (string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	s, err := secrets.Get(ctxChild, secretName, meta.GetOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	// Load `ca.crt` field
	cert, found := s.Data[constants.SecretCACertificate]
	if !found {
		return "", errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretCACertificate, secretName))
	}
	return string(cert), nil
}

// GetCASecret loads a secret with given name in the given namespace
// and extracts the `ca.crt` & `ca.key` field.
// If the secret does not exists or one of the fields is missing,
// an error is returned.
// Returns: certificate, private-key, isOwnedByDeployment, error
func GetCASecret(ctx context.Context, secrets secretv1.ReadInterface, secretName string,
	ownerRef *meta.OwnerReference) (string, string, bool, error) {
	s, err := secrets.Get(ctx, secretName, meta.GetOptions{})
	if err != nil {
		return "", "", false, errors.WithStack(err)
	}
	return GetCAFromSecret(s, ownerRef)
}

func GetCAFromSecret(s *core.Secret, ownerRef *meta.OwnerReference) (string, string, bool, error) {
	isOwned := false
	if ownerRef != nil {
		for _, x := range s.GetOwnerReferences() {
			if x.UID == ownerRef.UID {
				isOwned = true
				break
			}
		}
	}
	// Load `ca.crt` field
	cert, found := s.Data[constants.SecretCACertificate]
	if !found {
		return "", "", isOwned, errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretCACertificate, s.GetName()))
	}
	priv, found := s.Data[constants.SecretCAKey]
	if !found {
		return "", "", isOwned, errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretCAKey, s.GetName()))
	}
	return string(cert), string(priv), isOwned, nil
}

// CreateCASecret creates a secret used to store a PEM encoded CA certificate & private key.
func CreateCASecret(ctx context.Context, secrets secretv1.ModInterface, secretName string, certificate, key string,
	ownerRef *meta.OwnerReference) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretCACertificate: []byte(certificate),
			constants.SecretCAKey:         []byte(key),
		},
	}
	// Attach secret to owner
	AddOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(ctx, secret, meta.CreateOptions{}); err != nil {
		// Failed to create secret
		return kerrors.NewResourceError(err, secret)
	}
	return nil
}

// GetTLSKeyfileSecret loads a secret used to store a PEM encoded keyfile
// in the format ArangoDB accepts it for its `--ssl.keyfile` option.
// Returns: keyfile (pem encoded), error
func GetTLSKeyfileSecret(secrets secretv1.ReadInterface, secretName string) (string, error) {
	s, err := secrets.Get(context.Background(), secretName, meta.GetOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return GetTLSKeyfileFromSecret(s)
}

func GetTLSKeyfileFromSecret(s *core.Secret) (string, error) {
	// Load `tls.keyfile` field
	keyfile, found := s.Data[constants.SecretTLSKeyfile]
	if !found {
		return "", errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretTLSKeyfile, s.GetName()))
	}
	return string(keyfile), nil
}

// CreateTLSKeyfileSecret creates a secret used to store a PEM encoded keyfile
// in the format ArangoDB accepts it for its `--ssl.keyfile` option.
func CreateTLSKeyfileSecret(ctx context.Context, secrets secretv1.ModInterface, secretName string, keyfile string,
	ownerRef *meta.OwnerReference) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretTLSKeyfile: []byte(keyfile),
		},
	}
	// Attach secret to owner
	AddOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(ctx, secret, meta.CreateOptions{}); err != nil {
		// Failed to create secret
		return kerrors.NewResourceError(err, secret)
	}
	return nil
}

// ValidateTokenSecret checks that a secret with given name in given namespace
// exists and it contains a 'token' data field.
func ValidateTokenSecret(ctx context.Context, secrets secretv1.ReadInterface, secretName string) error {
	s, err := secrets.Get(ctx, secretName, meta.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return ValidateTokenFromSecret(s)
}

func ValidateTokenFromSecret(s *core.Secret) error {
	// Check `token` field
	_, found := s.Data[constants.SecretKeyToken]
	if !found {
		return errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretKeyToken, s.GetName()))
	}
	return nil
}

// GetTokenSecret loads the token secret from a Secret with given name.
func GetTokenSecret(ctx context.Context, secrets secretv1.ReadInterface, secretName string) (string, error) {
	s, err := secrets.Get(ctx, secretName, meta.GetOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return GetTokenFromSecret(s)
}

// GetTokenFromSecret loads the token secret from a Secret with given name.
func GetTokenFromSecret(s *core.Secret) (string, error) {
	// Take the first data from the token key
	data, found := s.Data[constants.SecretKeyToken]
	if !found {
		return "", errors.WithStack(errors.Newf("No '%s' data found in secret '%s'", constants.SecretKeyToken, s.GetName()))
	}
	return string(data), nil
}

// CreateTokenSecret creates a secret with given name in given namespace
// with a given token as value.
func CreateTokenSecret(ctx context.Context, secrets secretv1.ModInterface, secretName, token string,
	ownerRef *meta.OwnerReference) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretKeyToken: []byte(token),
		},
	}
	// Attach secret to owner
	AddOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(ctx, secret, meta.CreateOptions{}); err != nil {
		// Failed to create secret
		return kerrors.NewResourceError(err, secret)
	}
	return nil
}

// CreateJWTTokenFromSecret creates a JWT token
func CreateJWTTokenFromSecret(secret string, claims map[string]interface{}) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jg.NewWithClaims(jg.SigningMethodHS256, jg.MapClaims(claims))

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.WithStack(err)
	}

	return signedToken, nil
}

// CreateJWTFromSecret creates a JWT using the secret stored in secretSecretName and stores the
// result in a new secret called tokenSecretName
func CreateJWTFromSecret(ctx context.Context, cachedSecrets secretv1.ReadInterface, secrets secretv1.ModInterface, tokenSecretName, secretSecretName string, claims map[string]interface{}, ownerRef *meta.OwnerReference) error {
	secret, err := GetTokenSecret(ctx, cachedSecrets, secretSecretName)
	if err != nil {
		return errors.WithStack(err)
	}
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jg.NewWithClaims(jg.SigningMethodHS256, jg.MapClaims(claims))

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return errors.WithStack(err)
	}

	return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return CreateTokenSecret(ctxChild, secrets, tokenSecretName, signedToken, ownerRef)
	})
}

// CreateBasicAuthSecret creates a secret with given name in given namespace
// with a given username and password as value.
func CreateBasicAuthSecret(ctx context.Context, secrets secretv1.ModInterface, secretName, username, password string,
	ownerRef *meta.OwnerReference) error {
	// Create secret
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretUsername: []byte(username),
			constants.SecretPassword: []byte(password),
		},
	}
	// Attach secret to owner
	AddOwnerRefToObject(secret, ownerRef)
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := secrets.Create(ctxChild, secret, meta.CreateOptions{})
		return kerrors.NewResourceError(err, secret)
	})
	if err != nil {
		// Failed to create secret
		return errors.WithStack(err)
	}
	return nil
}

// GetBasicAuthSecret loads a secret with given name in the given namespace
// and extracts the `username` & `password` field.
// If the secret does not exists or one of the fields is missing,
// an error is returned.
// Returns: username, password, error
func GetBasicAuthSecret(secrets secretv1.Interface, secretName string) (string, string, error) {
	s, err := secrets.Get(context.Background(), secretName, meta.GetOptions{})
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	return GetSecretAuthCredentials(s)
}

// GetSecretAuthCredentials returns username and password from the secret
func GetSecretAuthCredentials(secret *core.Secret) (string, string, error) {
	username, found := secret.Data[constants.SecretUsername]
	if !found {
		return "", "", errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretUsername, secret.Name))
	}
	password, found := secret.Data[constants.SecretPassword]
	if !found {
		return "", "", errors.WithStack(errors.Newf("No '%s' found in secret '%s'", constants.SecretPassword, secret.Name))
	}
	return string(username), string(password), nil
}
