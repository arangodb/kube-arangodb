//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package log

import "github.com/rs/zerolog"

type Factory interface {
	Info() *zerolog.Event

	Str(key, value string) Factory
}

func NewFactory(logger zerolog.Logger) Factory {
	return &factory{
		logger: logger,
	}
}

type factory struct {
	logger zerolog.Logger
}

func (f factory) Info() *zerolog.Event {
	return f.logger.Info()
}

func (f factory) Str(key, value string) Factory {
	return NewFactory(f.logger.With().Str(key, value).Logger())
}
