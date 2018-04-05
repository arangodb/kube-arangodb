//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	certificates "github.com/arangodb-helper/go-certificates"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Config settings for the Server
type Config struct {
	Address            string // Address to listen on
	TLSSecretName      string // Name of secret containing TLS certificate
	TLSSecretNamespace string // Namespace of secret containing TLS certificate
	PodName            string // Name of the Pod we're running in
	PodIP              string // IP address of the Pod we're running in
}

// Server is the HTTPS server for the operator.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new server, fetching/preparing a TLS certificate.
func NewServer(cli corev1.CoreV1Interface, handler http.Handler, cfg Config) (*Server, error) {
	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	var cert, key string
	if cfg.TLSSecretName != "" && cfg.TLSSecretNamespace != "" {
		// Load TLS certificate from secret
		s, err := cli.Secrets(cfg.TLSSecretNamespace).Get(cfg.TLSSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, maskAny(err)
		}
		certBytes, found := s.Data[v1.TLSCertKey]
		if !found {
			return nil, maskAny(fmt.Errorf("No %s found in secret %s", v1.TLSCertKey, cfg.TLSSecretName))
		}
		keyBytes, found := s.Data[v1.TLSPrivateKeyKey]
		if !found {
			return nil, maskAny(fmt.Errorf("No %s found in secret %s", v1.TLSPrivateKeyKey, cfg.TLSSecretName))
		}
		cert = string(certBytes)
		key = string(keyBytes)
	} else {
		// Secret not specified, create our own TLS certificate
		options := certificates.CreateCertificateOptions{
			CommonName: cfg.PodName,
			Hosts:      []string{cfg.PodName, cfg.PodIP},
			ValidFrom:  time.Now(),
			ValidFor:   time.Hour * 24 * 365 * 10,
			IsCA:       false,
			ECDSACurve: "P256",
		}
		var err error
		cert, key, err = certificates.CreateCertificate(options, nil)
		if err != nil {
			return nil, maskAny(err)
		}
	}
	tlsConfig, err := createTLSConfig(cert, key)
	if err != nil {
		return nil, maskAny(err)
	}
	tlsConfig.BuildNameToCertificate()
	httpServer.TLSConfig = tlsConfig

	return &Server{
		httpServer: httpServer,
	}, nil
}

// Run the server until the program stops.
func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return maskAny(err)
	}
	return nil
}

// createTLSConfig creates a TLS config based on given config
func createTLSConfig(cert, key string) (*tls.Config, error) {
	var result *tls.Config
	c, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, maskAny(err)
	}
	result = &tls.Config{
		Certificates: []tls.Certificate{c},
	}
	return result, nil
}
