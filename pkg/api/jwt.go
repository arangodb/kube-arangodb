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

package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	jg "github.com/golang-jwt/jwt"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	secret "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// ensureJWT ensure that JWT signing key exists or creates a new one.
// It also saves new token into secret if it is not present.
// Returns JWT signing key.
func ensureJWT(cli typedCore.CoreV1Interface, cfg ServerConfig) (string, error) {
	secrets := cli.Secrets(cfg.Namespace)

	signingKey, err := k8sutil.GetTokenSecret(context.Background(), secrets, cfg.JWTKeySecretName)
	if err != nil && kerrors.IsNotFound(err) || signingKey == "" {
		signingKey, err = createSigningKey(secrets, cfg.JWTKeySecretName)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", errors.WithStack(err)
	}

	_, err = k8sutil.GetTokenSecret(context.Background(), secrets, cfg.JWTSecretName)
	if err != nil && kerrors.IsNotFound(err) {
		err = generateAndSaveJWT(secrets, cfg)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", errors.WithStack(err)
	}
	return signingKey, nil
}

// generateAndSaveJWT tries to generate new JWT using signing key retrieved from secret.
// If it is not present, it creates a new key.
// The resulting JWT is stored in secrets.
func generateAndSaveJWT(secrets secret.Interface, cfg ServerConfig) error {
	claims := jg.MapClaims{
		"iss": fmt.Sprintf("kube-arangodb/%s", cfg.ServerName),
		"iat": time.Now().Unix(),
	}
	err := k8sutil.CreateJWTFromSecret(context.Background(), secrets, secrets, cfg.JWTSecretName, cfg.JWTKeySecretName, claims, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	return err
}

func createSigningKey(secrets secret.ModInterface, keySecretName string) (string, error) {
	signingKey := make([]byte, 32)
	_, err := rand.Read(signingKey)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(context.Background(), func(ctxChild context.Context) error {
		return k8sutil.CreateTokenSecret(ctxChild, secrets, keySecretName, string(signingKey), nil)
	})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(signingKey), nil
}
