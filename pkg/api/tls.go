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
	"crypto/tls"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb-helper/go-certificates"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func prepareTLSConfig(cli typedCore.CoreV1Interface, cfg ServerConfig) (*tls.Config, error) {
	cert, key, err := loadOrSelfSignCertificate(cli, cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig, err := createTLSConfig(cert, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return tlsConfig, nil
}

// loadOrSelfSignCertificate loads TLS certificate from secret or creates a new one
func loadOrSelfSignCertificate(cli typedCore.CoreV1Interface, cfg ServerConfig) (string, string, error) {
	if cfg.TLSSecretName != "" {
		// Load TLS certificate from secret
		s, err := cli.Secrets(cfg.Namespace).Get(context.Background(), cfg.TLSSecretName, meta.GetOptions{})
		if err != nil {
			return "", "", err
		}
		certBytes, found := s.Data[core.TLSCertKey]
		if !found {
			return "", "", errors.Newf("No %s found in secret %s", core.TLSCertKey, cfg.TLSSecretName)
		}
		keyBytes, found := s.Data[core.TLSPrivateKeyKey]
		if !found {
			return "", "", errors.Newf("No %s found in secret %s", core.TLSPrivateKeyKey, cfg.TLSSecretName)
		}
		return string(certBytes), string(keyBytes), nil
	}
	// Secret not specified, create our own TLS certificate
	options := certificates.CreateCertificateOptions{
		CommonName: cfg.ServerName,
		Hosts:      append([]string{cfg.ServerName}, cfg.ServerAltNames...),
		ValidFrom:  time.Now(),
		ValidFor:   time.Hour * 24 * 365 * 10,
		IsCA:       false,
		ECDSACurve: "P256",
	}
	return certificates.CreateCertificate(options, nil)
}

// createTLSConfig creates a TLS config based on given config
func createTLSConfig(cert, key string) (*tls.Config, error) {
	var result *tls.Config
	c, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result = &tls.Config{
		Certificates: []tls.Certificate{c},
	}
	return result, nil
}
