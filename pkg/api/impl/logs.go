//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package impl

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pb "github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var loglevelMap = map[pb.LogLevel]logging.Level{
	pb.LogLevel_LOG_LEVEL_TRACE_UNSPECIFIED: logging.Trace,
	pb.LogLevel_LOG_LEVEL_DEBUG:             logging.Debug,
	pb.LogLevel_LOG_LEVEL_INFO:              logging.Info,
	pb.LogLevel_LOG_LEVEL_WARN:              logging.Warn,
	pb.LogLevel_LOG_LEVEL_ERROR:             logging.Error,
	pb.LogLevel_LOG_LEVEL_FATAL:             logging.Fatal,
}

func logLevelToGRPC(l logging.Level) pb.LogLevel {
	for grpcVal, localVal := range loglevelMap {
		if l == localVal {
			return grpcVal
		}
	}
	return pb.LogLevel_LOG_LEVEL_DEBUG
}

func (i *implementation) getLogLevelsByTopics() map[string]logging.Level {
	return logging.Global().LogLevels()
}

func (i *implementation) setLogLevelsByTopics(logLevels map[string]logging.Level) {
	logging.Global().ApplyLogLevels(logLevels)
}

func (i *implementation) GetLogLevel(ctx context.Context, _ *pbSharedV1.Empty) (*pb.LogLevelConfig, error) {
	if i.authenticate(ctx) != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	l := i.getLogLevelsByTopics()

	topics := make(map[string]pb.LogLevel, len(l))
	for topic, level := range l {
		topics[topic] = logLevelToGRPC(level)
	}
	return &pb.LogLevelConfig{
		Topics: topics,
	}, nil
}

func (i *implementation) SetLogLevel(ctx context.Context, cfg *pb.LogLevelConfig) (*pbSharedV1.Empty, error) {
	if i.authenticate(ctx) != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	l := make(map[string]logging.Level, len(cfg.Topics))
	for topic, grpcLevel := range cfg.Topics {
		level, ok := loglevelMap[grpcLevel]
		if !ok {
			return &pbSharedV1.Empty{}, status.Error(codes.NotFound, fmt.Sprintf("unknown log level %s for topic %s", grpcLevel, topic))
		}
		l[topic] = level
	}
	i.setLogLevelsByTopics(l)
	return &pbSharedV1.Empty{}, nil
}
