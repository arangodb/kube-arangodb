//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/proto"
)

func outgoingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	// Pass set-cookie as it is
	case "set-cookie":
		return key, true
	case "location":
		return "Location", true
	default:
		return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
	}
}

func forwardResponseOption(ctx context.Context, w http.ResponseWriter, message proto.Message) error {
	headers := w.Header()
	if _, ok := headers["Location"]; ok {
		w.WriteHeader(http.StatusFound)
	}

	return nil
}
