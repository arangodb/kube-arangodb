//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"context"
	"crypto/tls"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb-helper/go-certificates"

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

func NewStatisTLSConfig(cfg *tls.Config) TLSConfigFetcher {
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
