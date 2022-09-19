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
	"net/http"

	"github.com/gin-gonic/gin"
	prometheus "github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

func buildHTTPHandler(s *Server, cfg ServerConfig, auth *authorization) (http.Handler, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	versionV1Responder, err := operatorHTTP.NewSimpleJSONResponse(version.GetVersionV1())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r.GET("/_api/version", gin.WrapF(versionV1Responder.ServeHTTP))
	r.GET("/api/v1/version", gin.WrapF(versionV1Responder.ServeHTTP))
	r.GET("/health", gin.WrapF(cfg.LivelinessProbe.LivenessHandler))

	var readyProbes []*probe.ReadyProbe
	if cfg.ProbeDeployment.Enabled {
		r.GET("/ready/deployment", gin.WrapF(cfg.ProbeDeployment.Probe.ReadyHandler))
		readyProbes = append(readyProbes, cfg.ProbeDeployment.Probe)
	}
	if cfg.ProbeDeploymentReplication.Enabled {
		r.GET("/ready/deployment-replication", gin.WrapF(cfg.ProbeDeploymentReplication.Probe.ReadyHandler))
		readyProbes = append(readyProbes, cfg.ProbeDeploymentReplication.Probe)
	}
	if cfg.ProbeStorage.Enabled {
		r.GET("/ready/storage", gin.WrapF(cfg.ProbeStorage.Probe.ReadyHandler))
		readyProbes = append(readyProbes, cfg.ProbeStorage.Probe)
	}
	r.GET("/ready", gin.WrapF(handleGetReady(readyProbes...)))

	r.GET("/metrics", auth.ensureHTTPAuth, gin.WrapH(prometheus.Handler()))
	r.GET("/log/level", auth.ensureHTTPAuth, s.handleGetLogLevel)
	r.POST("/log/level", auth.ensureHTTPAuth, s.handlePostLogLevel)

	return r, nil
}

func handleGetReady(probes ...*probe.ReadyProbe) func(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleGetLogLevel(c *gin.Context) {
	logLevels := s.getLogLevelsByTopics()
	topics := make(map[string]string, len(logLevels))
	for topic, level := range logLevels {
		topics[topic] = level.String()
	}
	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
	})
}

func (s *Server) handlePostLogLevel(c *gin.Context) {
	var req = struct {
		Topics map[string]string `json:"topics"`
	}{}

	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	logLevels := make(map[string]logging.Level, len(req.Topics))
	for topic, levelStr := range req.Topics {
		l, err := logging.ParseLogLevel(levelStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
			return
		}
		logLevels[topic] = l
	}

	s.setLogLevelsByTopics(logLevels)
	c.JSON(http.StatusOK, gin.H{})
}
