//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiff(t *testing.T) {
	type args struct {
		compareWhat []string
		compareTo   []string
	}
	tests := map[string]struct {
		args args
		want []string
	}{
		"two nil slices": {},
		"source slice is nil": {
			args: args{
				compareWhat: nil,
				compareTo:   []string{"1"},
			},
			want: nil,
		},
		"destination slice is nil": {
			args: args{
				compareWhat: []string{"1"},
			},
			want: []string{"1"},
		},
		"source slice has more elements": {
			args: args{
				compareWhat: []string{"1", "2"},
				compareTo:   []string{"2"},
			},
			want: []string{"1"},
		},
		"destination slice has more elements": {
			args: args{
				compareWhat: []string{"1", "2"},
				compareTo:   []string{"1", "2", "3"},
			},
		},
		"destination and source slices have overlapping elements": {
			args: args{
				compareWhat: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				compareTo:   []string{"1", "3", "5", "7", "9"},
			},
			want: []string{"2", "4", "6", "8", "10"},
		},
		"destination slice contains source slice": {
			args: args{
				compareWhat: []string{"1", "3", "5", "7", "9"},
				compareTo:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			diff := Diff(testCase.args.compareWhat, testCase.args.compareTo)
			assert.Equal(t, testCase.want, diff)
		})
	}
}
