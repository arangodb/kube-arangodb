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
	"fmt"

	pb "github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

func (s *Server) GetVersion(ctx context.Context, _ *pb.Empty) (*pb.Version, error) {
	v := version.GetVersionV1()
	return &pb.Version{
		Version:   string(v.Version),
		Build:     v.Build,
		Edition:   string(v.Edition),
		GoVersion: v.GoVersion,
		BuildDate: v.BuildDate,
	}, nil
}

var loglevelMap = map[pb.LogLevel]logging.Level{
	pb.LogLevel_TRACE: logging.Trace,
	pb.LogLevel_DEBUG: logging.Debug,
	pb.LogLevel_INFO:  logging.Info,
	pb.LogLevel_WARN:  logging.Warn,
	pb.LogLevel_ERROR: logging.Error,
	pb.LogLevel_FATAL: logging.Fatal,
}

func logLevelToGRPC(l logging.Level) pb.LogLevel {
	for grpcVal, localVal := range loglevelMap {
		if l == localVal {
			return grpcVal
		}
	}
	return pb.LogLevel_DEBUG
}

func (s *Server) GetLogLevel(ctx context.Context, _ *pb.Empty) (*pb.LogLevelConfig, error) {
	l := s.getLogLevelsByTopics()

	topics := make(map[string]pb.LogLevel, len(l))
	for topic, level := range l {
		topics[topic] = logLevelToGRPC(level)
	}
	return &pb.LogLevelConfig{
		Topics: topics,
	}, nil
}

func (s *Server) SetLogLevel(ctx context.Context, cfg *pb.LogLevelConfig) (*pb.Empty, error) {
	l := make(map[string]logging.Level, len(cfg.Topics))
	for topic, grpcLevel := range cfg.Topics {
		level, ok := loglevelMap[grpcLevel]
		if !ok {
			return &pb.Empty{}, fmt.Errorf("unknown log level %s for topic %s", grpcLevel, topic)
		}
		l[topic] = level
	}
	s.setLogLevelsByTopics(l)
	return &pb.Empty{}, nil
}
