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

package client

import (
	"time"

	"github.com/arangodb/go-driver/util/connection/wrappers"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func newClientLogger(log logging.Logger) wrappers.Logger {
	return logger{log: log}
}

type logger struct {
	log logging.Logger
}

func (l logger) Int(key string, value int) wrappers.Event {
	return logger{log: l.log.Int(key, value)}
}

func (l logger) Str(key, value string) wrappers.Event {
	return logger{log: l.log.Str(key, value)}
}

func (l logger) Time(key string, value time.Time) wrappers.Event {
	return logger{log: l.log.Time(key, value)}
}

func (l logger) Duration(key string, value time.Duration) wrappers.Event {
	return logger{log: l.log.Dur(key, value)}
}

func (l logger) Interface(key string, value interface{}) wrappers.Event {
	return logger{log: l.log.Interface(key, value)}
}

func (l logger) Msgf(format string, args ...interface{}) {
	l.log.Trace(format, args...)
}

func (l logger) Log() wrappers.Event {
	return logger{log: l.log}
}
