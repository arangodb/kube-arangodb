//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	registerFeature(gracefulShutdown)
	registerFeature(optionalGracefulShutdown)
}

var gracefulShutdown = &feature{
	name:               "graceful-shutdown",
	description:        "Define graceful shutdown, using finalizers, is enabled",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             true,
}

var optionalGracefulShutdown = &feature{
	name:               "optional-graceful-shutdown",
	description:        "Define graceful shutdown, using finalizers, is optional and can fail in case of connection issues",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   false,
	hidden:             true,
}

func GracefulShutdown() Feature {
	return gracefulShutdown
}

func OptionalGracefulShutdown() Feature {
	return optionalGracefulShutdown
}
