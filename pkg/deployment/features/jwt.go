//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package features

func init() {
	registerFeature(jwtRotation)
	registerFeature(jwtAsymmetricKey)
}

var jwtRotation = &feature{
	name:               "jwt-rotation",
	description:        "JWT Token rotation in runtime",
	enterpriseRequired: true,
	enabledByDefault:   true,
	hidden:             true,
}

var jwtAsymmetricKey = &feature{
	name:               "jwt-asymmetric-key",
	description:        "Uses Asymmetric Key as a default in ArangoDB",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             true,
	version:            newFeatureVersion("3.12.8", NoVersionLimit),
}

func JWTRotation() Feature {
	return jwtRotation
}

func JWTAsymmetricKey() Feature {
	return jwtAsymmetricKey
}
