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

const (
	LoggerNameOperator              = "operator"
	LoggerNameDeployment            = "deployment"
	LoggerNameInspector             = "inspector"
	LoggerNameKLog                  = "klog"
	LoggerNameServer                = "server"
	LoggerNameDeploymentReplication = "deployment-replication"
	LoggerNameStorage               = "storage"
	LoggerNameProvisioner           = "provisioner"
	LoggerNameReconciliation        = "reconciliation"
	LoggerNameEventRecorder         = "event-recorder"
)

var defaultLogLevels = map[string]zerolog.Level{
	LoggerNameInspector: zerolog.WarnLevel,
}

func LoggerNames() []string {
	return []string{
		LoggerNameOperator,
		LoggerNameDeployment,
		LoggerNameInspector,
		LoggerNameKLog,
		LoggerNameServer,
		LoggerNameDeploymentReplication,
		LoggerNameStorage,
		LoggerNameProvisioner,
		LoggerNameReconciliation,
		LoggerNameEventRecorder,
	}
}
