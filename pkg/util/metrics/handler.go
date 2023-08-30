//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package metrics

import (
	"net/http"

	prometheus "github.com/prometheus/client_golang/prometheus/promhttp"

	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

const MaxMetricsBufferedSize = 1024 * 1024

func Handler() http.HandlerFunc {
	return operatorHTTP.WithTextContentType(operatorHTTP.WithNoContent(operatorHTTP.WithBuffer(MaxMetricsBufferedSize, operatorHTTP.WithEncoding(prometheus.Handler().ServeHTTP))))
}
