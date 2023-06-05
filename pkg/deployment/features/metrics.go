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

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func init() {
	registerFeature(metricsExporter)
}

var metricsExporter = &feature{
	name:               "metrics-exporter",
	description:        "Define if internal metrics-exporter should be used",
	version:            "3.6.0",
	enterpriseRequired: false,
	enabledByDefault:   true,
	deprecated:         "It is always set to True",
	constValue:         util.NewType[bool](true),
	hidden:             true,
}

// deprecated
func MetricsExporter() Feature {
	return metricsExporter
}
