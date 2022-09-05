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

import (
	"strings"

	"github.com/rs/zerolog"
)

func ParseLogLevelsFromArgs(in []string) (map[string]Level, error) {
	r := make(map[string]Level)

	for _, level := range in {
		z := strings.SplitN(level, "=", 2)

		switch len(z) {
		case 1:
			l, err := ParseLogLevel(z[0])
			if err != nil {
				return nil, err
			}

			r[TopicAll] = l
		case 2:
			l, err := ParseLogLevel(z[1])
			if err != nil {
				return nil, err
			}

			r[z[0]] = l
		}
	}

	return r, nil
}

func ParseLogLevel(in string) (Level, error) {
	l, err := zerolog.ParseLevel(strings.ToLower(in))
	if err != nil {
		return Debug, err
	}
	return Level(l), nil
}
