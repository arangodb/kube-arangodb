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

package resources

import core "k8s.io/api/core/v1"

// NewEnvBuilder creates new ENV builder
func NewEnvBuilder() EnvBuilder {
	return make(EnvBuilder, 0)
}

// EnvBuilder build environment variables
type EnvBuilder []core.EnvVar

// Add append or override flag in envs. Flag is value was modified is returned
func (e *EnvBuilder) Add(override bool, envs ...core.EnvVar) (modified bool) {
	for _, env := range envs {
		if id, ok := e.getID(env); ok {
			if override {
				(*e)[id] = env
			}
		}

		*e = append(*e, env)
		modified = true
	}

	return
}

func (e *EnvBuilder) getID(env core.EnvVar) (int, bool) {
	for id, currentEnvs := range *e {
		if currentEnvs.Name == env.Name {
			return id, true
		}
	}

	return -1, false
}

// GetEnvList return copy of env list
func (e EnvBuilder) GetEnvList() []core.EnvVar {
	if len(e) == 0 {
		return nil
	}

	l := make([]core.EnvVar, len(e))

	for id, env := range e {
		l[id] = *env.DeepCopy()
	}

	return l
}
