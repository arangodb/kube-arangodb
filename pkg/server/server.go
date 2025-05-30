//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package server

import (
	"context"
	"crypto/tls"
	goHttp "net/http"
	goStrings "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-assets"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/dashboard"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

// Config settings for the Server
type Config struct {
	Namespace          string
	Address            string // Address to listen on
	TLSSecretName      string // Name of secret containing TLS certificate
	TLSSecretNamespace string // Namespace of secret containing TLS certificate
	PodName            string // Name of the Pod we're running in
	PodIP              string // IP address of the Pod we're running in
	AdminSecretName    string // Name of basic authentication secret containing the admin username+password of the dashboard
	AllowAnonymous     bool   // If set, anonymous access to dashboard is allowed
}

type OperatorDependency struct {
	Enabled bool
	Probe   *probe.ReadyProbe
}

// Dependencies of the Server
type Dependencies struct {
	LivenessProbe         *probe.LivenessProbe
	Deployment            OperatorDependency
	DeploymentReplication OperatorDependency
	Storage               OperatorDependency
	Backup                OperatorDependency
	Apps                  OperatorDependency
	ML                    OperatorDependency
	Analytics             OperatorDependency
	Networking            OperatorDependency
	Platform              OperatorDependency
	Scheduler             OperatorDependency
	ClusterSync           OperatorDependency
	Operators             Operators
	Secrets               typedCore.SecretInterface
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
	cfg        Config
	deps       Dependencies
	httpServer *goHttp.Server
	auth       *serverAuthentication
}

// NewServer creates a new server, fetching/preparing a TLS certificate.
func NewServer(cli typedCore.CoreV1Interface, cfg Config, deps Dependencies) (*Server, error) {
	httpServer := &goHttp.Server{
		Addr:              cfg.Address,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		TLSNextProto:      make(map[string]func(*goHttp.Server, *tls.Conn, goHttp.Handler)),
	}

	var fetcher util.TLSConfigFetcher
	if cfg.TLSSecretName != "" && cfg.TLSSecretNamespace != "" {
		serverLogger.Str("addr", cfg.Address).Str("secret", cfg.TLSSecretName).Str("secret-namespace", cfg.TLSSecretNamespace).Info("Using existing TLS Certificate")
		fetcher = util.NewSecretTLSConfig(cli.Secrets(cfg.TLSSecretNamespace), cfg.TLSSecretName)
	} else {
		serverLogger.Str("addr", cfg.Address).Info("Using SelfSigned TLS Certificate")
		fetcher = util.NewSelfSignedTLSConfig(cfg.PodName, cfg.PodIP)
	}

	tlsConfig, err := fetcher.Eval(context.Background())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	httpServer.TLSConfig = tlsConfig

	// Builder server
	s := &Server{
		cfg:        cfg,
		deps:       deps,
		httpServer: httpServer,
		auth:       newServerAuthentication(deps.Secrets, cfg.AdminSecretName, cfg.AllowAnonymous),
	}

	// Build router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", gin.WrapF(deps.LivenessProbe.LivenessHandler))

	versionV1Responser, err := operatorHTTP.NewSimpleJSONResponse(version.GetVersionV1())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r.GET("/_api/version", gin.WrapF(versionV1Responser.ServeHTTP))
	r.GET("/api/v1/version", gin.WrapF(versionV1Responser.ServeHTTP))

	var readyProbes []*probe.ReadyProbe
	if deps.Deployment.Enabled {
		r.GET("/ready/deployment", gin.WrapF(deps.Deployment.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Deployment.Probe)
	}
	if deps.DeploymentReplication.Enabled {
		r.GET("/ready/deployment-replication", gin.WrapF(deps.DeploymentReplication.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.DeploymentReplication.Probe)
	}
	if deps.Storage.Enabled {
		r.GET("/ready/storage", gin.WrapF(deps.Storage.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Storage.Probe)
	}
	if deps.Backup.Enabled {
		r.GET("/ready/backup", gin.WrapF(deps.Backup.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Backup.Probe)
	}
	if deps.Apps.Enabled {
		r.GET("/ready/apps", gin.WrapF(deps.Apps.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Apps.Probe)
	}
	if deps.ML.Enabled {
		r.GET("/ready/ml", gin.WrapF(deps.ML.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.ML.Probe)
	}
	if deps.Analytics.Enabled {
		r.GET("/ready/analytics", gin.WrapF(deps.Analytics.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Analytics.Probe)
	}
	if deps.Networking.Enabled {
		r.GET("/ready/networking", gin.WrapF(deps.Networking.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Networking.Probe)
	}
	if deps.Platform.Enabled {
		r.GET("/ready/platform", gin.WrapF(deps.Platform.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Platform.Probe)
	}
	if deps.Scheduler.Enabled {
		r.GET("/ready/scheduler", gin.WrapF(deps.Scheduler.Probe.ReadyHandler))
		readyProbes = append(readyProbes, deps.Scheduler.Probe)
	}
	r.GET("/ready", gin.WrapF(ready(readyProbes...)))
	r.GET("/metrics", gin.WrapF(metrics.Handler()))
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
		localPath := "/" + goStrings.TrimPrefix(path, "/")
		r.GET(localPath, createAssetFileHandler(file))
	}
	httpServer.Handler = r

	return s, nil
}

// createAssetFileHandler creates a gin handler to serve the content
// of the given asset file.
func createAssetFileHandler(file *assets.File) func(c *gin.Context) {
	return func(c *gin.Context) {
		goHttp.ServeContent(c.Writer, c.Request, file.Name(), file.ModTime(), file)
	}
}

// Run the server until the program stops.
func (s *Server) Run() error {
	serverLogger.Info("Serving on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != goHttp.ErrServerClosed {
		return errors.WithStack(err)
	}
	return nil
}

func ready(probes ...*probe.ReadyProbe) func(w goHttp.ResponseWriter, r *goHttp.Request) {
	return func(w goHttp.ResponseWriter, r *goHttp.Request) {
		for _, probe := range probes {
			if !probe.IsReady() {
				w.WriteHeader(goHttp.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(goHttp.StatusOK)
	}
}
