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
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	pb "github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
)

var apiLogger = logging.Global().RegisterAndGetLogger("api-server", logging.Info)

type Server struct {
	httpServer  *http.Server
	grpcServer  *grpc.Server
	grpcAddress string

	pb.UnimplementedOperatorServer
}

type ReadinessProbeConfig struct {
	Enabled bool
	Probe   *probe.ReadyProbe
}

// ServerConfig settings for the Server
type ServerConfig struct {
	Namespace                  string
	ServerName                 string
	ServerAltNames             []string
	HTTPAddress                string
	GRPCAddress                string
	TLSSecretName              string
	JWTSecretName              string
	JWTKeySecretName           string
	LivelinessProbe            *probe.LivenessProbe
	ProbeDeployment            ReadinessProbeConfig
	ProbeDeploymentReplication ReadinessProbeConfig
	ProbeStorage               ReadinessProbeConfig
}

// NewServer creates and configure a new Server
func NewServer(cli typedCore.CoreV1Interface, cfg ServerConfig) (*Server, error) {
	jwtSigningKey, err := ensureJWT(cli, cfg)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := prepareTLSConfig(cli, cfg)
	if err != nil {
		return nil, err
	}

	auth := &authorization{jwtSigningKey: jwtSigningKey}

	s := &Server{
		httpServer: &http.Server{
			Addr:              cfg.HTTPAddress,
			ReadTimeout:       time.Second * 30,
			ReadHeaderTimeout: time.Second * 15,
			WriteTimeout:      time.Second * 30,
			TLSConfig:         tlsConfig,
		},
		grpcServer: grpc.NewServer(
			grpc.UnaryInterceptor(auth.ensureGRPCAuth),
			grpc.Creds(credentials.NewTLS(tlsConfig)),
		),
		grpcAddress: cfg.GRPCAddress,
	}
	handler, err := buildHTTPHandler(s, cfg, auth)
	if err != nil {
		return nil, err
	}
	s.httpServer.Handler = handler

	pb.RegisterOperatorServer(s.grpcServer, s)
	return s, nil
}

func (s *Server) Run() error {
	g := errgroup.Group{}
	g.Go(func() error {
		apiLogger.Info("Serving HTTP API on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		apiLogger.Info("Serving GRPC API on %s", s.grpcAddress)
		ln, err := net.Listen("tcp", s.grpcAddress)
		if err != nil {
			return err
		}
		defer ln.Close()

		if err := s.grpcServer.Serve(ln); err != nil && err != grpc.ErrServerStopped {
			return err
		}
		return nil
	})
	return g.Wait()
}
