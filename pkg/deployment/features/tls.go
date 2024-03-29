//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	registerFeature(tlsRotation)
	registerFeature(tlsSNI)
}

var tlsRotation Feature = &feature{
	name:               "tls-rotation",
	description:        "TLS Keyfile rotation in runtime",
	enterpriseRequired: true,
	enabledByDefault:   true,
	hidden:             true,
}

func TLSRotation() Feature {
	return tlsRotation
}

var tlsSNI Feature = &feature{
	name:               "tls-sni",
	description:        "TLS SNI Support",
	enterpriseRequired: true,
	enabledByDefault:   true,
}

func TLSSNI() Feature {
	return tlsSNI
}
