//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package utils

import "math/rand"

var (
	randomBase = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func RandomString(n int) string {
	return RandomStringFrom(n, randomBase)
}

func RandomStringFrom(n int, base []rune) string {
	runes := make([]rune, n)
	for id := range runes {
		runes[id] = base[rand.Intn(len(base))]
	}

	return string(runes)
}
