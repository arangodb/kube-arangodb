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

package logging

import "github.com/rs/zerolog"

type Level zerolog.Level

func (l Level) New() *Level {
	return &l
}

const (
	Trace = Level(zerolog.TraceLevel)
	Debug = Level(zerolog.DebugLevel)
	Info  = Level(zerolog.InfoLevel)
	Warn  = Level(zerolog.WarnLevel)
	Error = Level(zerolog.ErrorLevel)
	Fatal = Level(zerolog.FatalLevel)
)

func (l Level) String() string {
	return zerolog.Level(l).String()
}
