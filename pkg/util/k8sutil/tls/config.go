//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package tls

import (
	"context"
	"crypto/tls"
	"time"

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb-helper/go-certificates"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

type TLSConfigFetcher func(ctx context.Context) (*tls.Config, error)

func (t TLSConfigFetcher) Eval(ctx context.Context) (*tls.Config, error) {
	if t == nil {
		return EmptyTLSConfig(ctx)
	}

	return t(ctx)
}

func EmptyTLSConfig(ctx context.Context) (*tls.Config, error) {
	return nil, nil
}

func NewStaticTLSConfig(cfg *tls.Config) TLSConfigFetcher {
	return func(ctx context.Context) (*tls.Config, error) {
		return cfg, nil
	}
}

func NewSelfSignedTLSConfig(cn string, names ...string) TLSConfigFetcher {
	return func(ctx context.Context) (*tls.Config, error) {
		options := certificates.CreateCertificateOptions{
			CommonName: cn,
			Hosts:      append([]string{cn}, names...),
			ValidFrom:  time.Now(),
			ValidFor:   time.Hour * 24 * 365 * 10,
			IsCA:       false,
			ECDSACurve: "P256",
		}

		cert, priv, err := certificates.CreateCertificate(options, nil)
		if err != nil {
			return nil, err
		}

		var result *tls.Config
		c, err := tls.X509KeyPair([]byte(cert), []byte(priv))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = &tls.Config{
			Certificates: []tls.Certificate{c},
		}
		return result, nil
	}
}

func NewSecretTLSConfig(client generic.GetInterface[*core.Secret], name string) TLSConfigFetcher {
	return func(ctx context.Context) (*tls.Config, error) {
		nctx, cancel := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		s, err := client.Get(nctx, name, meta.GetOptions{})
		if err != nil {
			return nil, err
		}
		certBytes, found := s.Data[core.TLSCertKey]
		if !found {
			return nil, errors.Errorf("No %s found in secret %s", core.TLSCertKey, name)
		}
		keyBytes, found := s.Data[core.TLSPrivateKeyKey]
		if !found {
			return nil, errors.Errorf("No %s found in secret %s", core.TLSPrivateKeyKey, name)
		}

		var result *tls.Config
		c, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = &tls.Config{
			Certificates: []tls.Certificate{c},
		}
		return result, nil
	}
}

func NewLocalSecretTLSCAConfig(client generic.Client[*core.Secret], name, cn string, names ...string) TLSConfigFetcher {
	return func(ctx context.Context) (*tls.Config, error) {
		ca, _, err := GetOrCreateTLSCAConfig(ctx, client, name)
		if err != nil {
			return nil, err
		}

		options := certificates.CreateCertificateOptions{
			CommonName: cn,
			Hosts:      append([]string{cn}, names...),
			ValidFrom:  time.Now(),
			ValidFor:   time.Hour * 24 * 365,
			IsCA:       false,
			ECDSACurve: "P256",
		}

		cert, priv, err := certificates.CreateCertificate(options, ca)
		if err != nil {
			return nil, err
		}

		var result *tls.Config
		c, err := tls.X509KeyPair([]byte(cert), []byte(priv))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = &tls.Config{
			Certificates: []tls.Certificate{c},
		}
		return result, nil
	}
}

func GetOrCreateTLSCAConfig(ctx context.Context, client generic.Client[*core.Secret], name string) (*certificates.CA, []byte, error) {
	secret, err := client.Get(ctx, name, meta.GetOptions{})

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, nil, errors.WithStack(err)
		}

		cert, key, err := CreateTLSCACertificate("arangodb-operator")
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}

		secret, err = client.Create(ctx, &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
			Data: map[string][]byte{
				utilConstants.SecretCACertificate: []byte(cert),
				utilConstants.SecretCAKey:         []byte(key),
			},
		}, meta.CreateOptions{})
		if err != nil {
			if !apiErrors.IsConflict(err) {
				return nil, nil, errors.WithStack(err)
			}

			if secret, err = client.Get(ctx, name, meta.GetOptions{}); err != nil {
				return nil, nil, err
			}
		}
	}

	if secret == nil {
		return nil, nil, errors.Errorf("Secret %s not found", name)
	}

	cert, ok := secret.Data[utilConstants.SecretCACertificate]
	if !ok {
		return nil, nil, errors.Errorf("Secret %s not valid: Key %s not found", name, utilConstants.SecretCACertificate)
	}

	key, ok := secret.Data[utilConstants.SecretCAKey]
	if !ok {
		return nil, nil, errors.Errorf("Secret %s not valid: Key %s not found", name, utilConstants.SecretCAKey)
	}

	certObj, keyObj, err := certificates.LoadFromPEM(string(cert), string(key))
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return &certificates.CA{
		Certificate: certObj,
		PrivateKey:  keyObj,
	}, cert, nil
}

func NewKeyfileTLSConfig(keyfile string) TLSConfigFetcher {
	return func(ctx context.Context) (*tls.Config, error) {
		certificate, err := certificates.LoadKeyFile(keyfile)
		if err != nil {
			return nil, err
		}

		return &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}, nil
	}
}
