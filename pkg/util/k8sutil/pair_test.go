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

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStringToOptionPair(t *testing.T) {
	tests := map[string]struct {
		arg  string
		want OptionPair
	}{
		"empty argument": {},
		"not trimmed key argument on the left side": {
			arg: "           --log.level=requests=debug",
			want: OptionPair{
				Key:   "--log.level",
				Value: "requests=debug",
			},
		},
		"key argument not trimmed on the both sides": {
			arg: "           --log.level   =requests=debug",
			want: OptionPair{
				Key:   "--log.level",
				Value: "requests=debug",
			},
		},
		"key argument not trimmed on the both sides without value": {
			arg: "  --log.level  ",
			want: OptionPair{
				Key: "--log.level",
			},
		},
		"key argument not trimmed on the both sides without value with equal": {
			arg: "  --log.level =",
			want: OptionPair{
				Key: "--log.level",
			},
		},
		"key argument not trimmed on the right side with some value": {
			arg: "--log.level =  value   ",
			want: OptionPair{
				Key:   "--log.level",
				Value: "  value   ",
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got := ExtractStringToOptionPair(testCase.arg)
			assert.Equal(t, testCase.want, got)
		})
	}
}
