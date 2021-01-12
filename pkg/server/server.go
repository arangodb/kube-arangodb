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

package server

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb-helper/go-certificates"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-assets"
	prometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/dashboard"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
)

// Config settings for the Server
type Config struct {
	Namespace string
	Address   string // Address to listen on

	MetricsServer HTTPServerConfig
	HealthServer  HTTPServerConfig

	HTTPServerEnabled  bool   // Determines if http server is enabled
	TLSSecretName      string // Name of secret containing TLS certificate
	TLSSecretNamespace string // Namespace of secret containing TLS certificate
	PodName            string // Name of the Pod we're running in
	PodIP              string // IP address of the Pod we're running in
	AdminSecretName    string // Name of basic authentication secret containing the admin username+password of the dashboard
	AllowAnonymous     bool   // If set, anonymous access to dashboard is allowed
}

type HTTPServerConfig struct {
	Enabled            bool
	Address            string
	TLSSecretName      string // Name of secret containing TLS certificate
	TLSSecretNamespace string // Namespace of secret containing TLS certificate
}

type OperatorDependency struct {
	Enabled bool
	Probe   *probe.ReadyProbe
}

// Dependencies of the Server
type Dependencies struct {
	Log                   zerolog.Logger
	LivenessProbe         *probe.LivenessProbe
	Deployment            OperatorDependency
	DeploymentReplication OperatorDependency
	Storage               OperatorDependency
	Backup                OperatorDependency
	Operators             Operators
	Secrets               corev1.SecretInterface
}

// Operators is the API provided to the server for accessing the various operators.
type Operators interface {
	// Return the deployment operator (if any)
	DeploymentOperator() DeploymentOperator
	// Return the deployment replication operator (if any)
	DeploymentReplicationOperator() DeploymentReplicationOperator
	// Return the local storage operator (if any)
	StorageOperator() StorageOperator
	// FindOtherOperators looks up references to other operators in the same Kubernetes cluster.
	FindOtherOperators() []OperatorReference
}

// Server is the HTTPS server for the operator.
type Server struct {
	cfg                                             Config
	deps                                            Dependencies
	httpServer, metricsHTTPServer, healthHTTPServer *http.Server
	auth                                            *serverAuthentication
}

// NewServer creates a new server, fetching/preparing a TLS certificate.
func NewServer(cli corev1.CoreV1Interface, cfg Config, deps Dependencies) (*Server, error) {
	// Builder server
	s := &Server{
		cfg:  cfg,
		deps: deps,
		auth: newServerAuthentication(deps.Log, deps.Secrets, cfg.AdminSecretName, cfg.AllowAnonymous),
	}

	if http, err := newHttpDashboardServer(s, cli); err != nil {
		return nil, err
	} else {
		s.httpServer = http
	}

	if http, err := newHttpMetricsServer(s, cli); err != nil {
		return nil, err
	} else {
		s.metricsHTTPServer = http
	}

	if http, err := newHttpHealthServer(s, cli); err != nil {
		return nil, err
	} else {
		s.healthHTTPServer = http
	}

	return s, nil
}

func newHttpMetricsServer(s *Server, cli corev1.CoreV1Interface) (*http.Server, error) {
	server, err := newHttpServer(s, cli, s.cfg.MetricsServer)
	if err != nil || server == nil {
		return server, err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/metrics", gin.WrapH(prometheus.Handler()))
	server.Handler = r
	return server, nil
}

func newHttpHealthServer(s *Server, cli corev1.CoreV1Interface) (*http.Server, error) {
	server, err := newHttpServer(s, cli, s.cfg.HealthServer)
	if err != nil || server == nil {
		return server, err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", gin.WrapF(s.deps.LivenessProbe.LivenessHandler))

	var readyProbes []*probe.ReadyProbe
	if s.deps.Deployment.Enabled {
		r.GET("/ready/deployment", gin.WrapF(s.deps.Deployment.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.Deployment.Probe)
	}
	if s.deps.DeploymentReplication.Enabled {
		r.GET("/ready/deployment-replication", gin.WrapF(s.deps.DeploymentReplication.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.DeploymentReplication.Probe)
	}
	if s.deps.Storage.Enabled {
		r.GET("/ready/storage", gin.WrapF(s.deps.Storage.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.Storage.Probe)
	}
	r.GET("/ready", gin.WrapF(ready(readyProbes...)))

	server.Handler = r
	return server, nil
}

func newHttpServer(s *Server, cli corev1.CoreV1Interface, cfg HTTPServerConfig) (*http.Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	httpServer := &http.Server{
		Addr:              cfg.Address,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if cfg.TLSSecretName != "" && cfg.TLSSecretNamespace != "" {
		// Load TLS certificate from secret
		secret, err := cli.Secrets(cfg.TLSSecretNamespace).Get(cfg.TLSSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		certBytes, found := secret.Data[core.TLSCertKey]
		if !found {
			return nil, errors.WithStack(errors.Newf("No %s found in secret %s", core.TLSCertKey, cfg.TLSSecretName))
		}
		keyBytes, found := secret.Data[core.TLSPrivateKeyKey]
		if !found {
			return nil, errors.WithStack(errors.Newf("No %s found in secret %s", core.TLSPrivateKeyKey, cfg.TLSSecretName))
		}
		cert := string(certBytes)
		key := string(keyBytes)
		tlsConfig, err := createTLSConfig(cert, key)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tlsConfig.BuildNameToCertificate()
		httpServer.TLSConfig = tlsConfig
	}

	return httpServer, nil
}

func newHttpDashboardServer(s *Server, cli corev1.CoreV1Interface) (*http.Server, error) {
	if !s.cfg.HTTPServerEnabled {
		return nil, nil
	}
	httpServer := &http.Server{
		Addr:              s.cfg.Address,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	var cert, key string
	if s.cfg.TLSSecretName != "" && s.cfg.TLSSecretNamespace != "" {
		// Load TLS certificate from secret
		secret, err := cli.Secrets(s.cfg.TLSSecretNamespace).Get(s.cfg.TLSSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		certBytes, found := secret.Data[core.TLSCertKey]
		if !found {
			return nil, errors.WithStack(errors.Newf("No %s found in secret %s", core.TLSCertKey, s.cfg.TLSSecretName))
		}
		keyBytes, found := secret.Data[core.TLSPrivateKeyKey]
		if !found {
			return nil, errors.WithStack(errors.Newf("No %s found in secret %s", core.TLSPrivateKeyKey, s.cfg.TLSSecretName))
		}
		cert = string(certBytes)
		key = string(keyBytes)
	} else {
		// Secret not specified, create our own TLS certificate
		options := certificates.CreateCertificateOptions{
			CommonName: s.cfg.PodName,
			Hosts:      []string{s.cfg.PodName, s.cfg.PodIP},
			ValidFrom:  time.Now(),
			ValidFor:   time.Hour * 24 * 365 * 10,
			IsCA:       false,
			ECDSACurve: "P256",
		}
		var err error
		cert, key, err = certificates.CreateCertificate(options, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	tlsConfig, err := createTLSConfig(cert, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig.BuildNameToCertificate()
	httpServer.TLSConfig = tlsConfig

	// Build router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", gin.WrapF(s.deps.LivenessProbe.LivenessHandler))

	var readyProbes []*probe.ReadyProbe
	if s.deps.Deployment.Enabled {
		r.GET("/ready/deployment", gin.WrapF(s.deps.Deployment.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.Deployment.Probe)
	}
	if s.deps.DeploymentReplication.Enabled {
		r.GET("/ready/deployment-replication", gin.WrapF(s.deps.DeploymentReplication.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.DeploymentReplication.Probe)
	}
	if s.deps.Storage.Enabled {
		r.GET("/ready/storage", gin.WrapF(s.deps.Storage.Probe.ReadyHandler))
		readyProbes = append(readyProbes, s.deps.Storage.Probe)
	}
	r.GET("/ready", gin.WrapF(ready(readyProbes...)))
	r.GET("/metrics", gin.WrapH(prometheus.Handler()))
	r.POST("/login", s.auth.handleLogin)
	api := r.Group("/api", s.auth.checkAuthentication)
	{
		api.GET("/operators", s.handleGetOperators)

		// Deployment operator
		api.GET("/deployment", s.handleGetDeployments)
		api.GET("/deployment/:name", s.handleGetDeploymentDetails)

		// Deployment replication operator
		api.GET("/deployment-replication", s.handleGetDeploymentReplications)
		api.GET("/deployment-replication/:name", s.handleGetDeploymentReplicationDetails)

		// Local storage operator
		api.GET("/storage", s.handleGetLocalStorages)
		api.GET("/storage/:name", s.handleGetLocalStorageDetails)
	}
	// Dashboard
	r.GET("/", createAssetFileHandler(dashboard.Assets.Files["index.html"]))
	for path, file := range dashboard.Assets.Files {
		localPath := "/" + strings.TrimPrefix(path, "/")
		r.GET(localPath, createAssetFileHandler(file))
	}
	httpServer.Handler = r

	return httpServer, nil
}

// createAssetFileHandler creates a gin handler to serve the content
// of the given asset file.
func createAssetFileHandler(file *assets.File) func(c *gin.Context) {
	return func(c *gin.Context) {
		http.ServeContent(c.Writer, c.Request, file.Name(), file.ModTime(), file)
	}
}

// Run the server until the program stops.
func (s *Server) Run() error {
	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		if err := s.runServer("Dashboard", s.httpServer); err != nil {
			s.deps.Log.Error().Err(err).Msgf("Unable to start dashboard server")
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.runServer("Metrics", s.metricsHTTPServer); err != nil {
			s.deps.Log.Error().Err(err).Msgf("Unable to start metrics server")
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.runServer("Health", s.healthHTTPServer); err != nil {
			s.deps.Log.Error().Err(err).Msgf("Unable to start health server")
		}
	}()

	go func() {
		defer wg.Done()

		// Trap signal
		signals := make(chan os.Signal)
		defer close(signals)
		signal.Notify(signals, os.Interrupt, os.Kill)
		<-signals
	}()

	wg.Wait()

	return nil
}

func (s *Server) runServer(name string, server *http.Server) error {
	if server == nil {
		s.deps.Log.Info().Msgf("%s Server is disabled", name)
		return nil
	}
	if server.TLSConfig != nil {
		s.deps.Log.Info().Msgf("Serving TLS %s on %s", name, server.Addr)
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return errors.WithStack(err)
		}
	} else {
		s.deps.Log.Info().Msgf("Serving %s on %s", name, server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return errors.WithStack(err)
		}
	}
	s.deps.Log.Info().Msgf("Serving %s on %s Done", name, server.Addr)
	return nil
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

func ready(probes ...*probe.ReadyProbe) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, probe := range probes {
			if !probe.IsReady() {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
