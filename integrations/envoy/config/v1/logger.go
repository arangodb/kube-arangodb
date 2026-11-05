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

import "github.com/arangodb/kube-arangodb/pkg/logging"

var logger = logging.Global().RegisterAndGetLogger("integration-envoy-config-v1", logging.Info)

// gcpLogger adapts our logger to the go-control-plane cache/server log.Logger interface.
type gcpLogger struct{}

func (gcpLogger) Debugf(format string, args ...interface{}) { logger.Debug(format, args...) }
func (gcpLogger) Infof(format string, args ...interface{})  { logger.Debug(format, args...) }
func (gcpLogger) Warnf(format string, args ...interface{})  { logger.Warn(format, args...) }
func (gcpLogger) Errorf(format string, args ...interface{}) { logger.Warn(format, args...) }
