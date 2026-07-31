//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

const (
	handlerHeartbeatInterval = 30 * time.Second
	handlerTTL               = time.Minute
)

var logger = logging.Global().RegisterAndGetLogger("connector-v1", logging.Info)

// startHeartbeat registers the handler in MetaStore and renews the entry every 30 seconds with a 1 minute TTL.
// Blocks until ctx is cancelled.
func startHeartbeat(ctx context.Context, meta pbMetaV1.MetaV1Client, linkID, handlerID string) {
	key := handlerKey(linkID, handlerID)

	register := func() {
		obj, err := anypb.New(timestamppb.Now())
		if err != nil {
			logger.Err(err).Warn("Failed to marshal heartbeat")
			return
		}

		_, err = meta.Set(ctx, &pbMetaV1.SetRequest{
			Key:    key,
			Object: obj,
			Ttl:    durationpb.New(handlerTTL),
		})
		if err != nil {
			logger.Err(err).Str("handler", handlerID).Warn("Failed to register handler heartbeat")
			return
		}

		logger.Str("handler", handlerID).Debug("Handler heartbeat registered")
	}

	// Initial registration
	register()

	ticker := time.NewTicker(handlerHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			register()
		}
	}
}
