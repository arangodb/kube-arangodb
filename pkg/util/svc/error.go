//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package svc

import "context"

type serviceError struct {
	error
}

func (p serviceError) StartWithHealth(ctx context.Context, health Health) ServiceStarter {
	return p
}

func (p serviceError) Address() string {
	return ""
}

func (p serviceError) Wait() error {
	return p
}

func (p serviceError) Update(key string, state HealthState) {

}

func (p serviceError) Start(ctx context.Context) ServiceStarter {
	return p
}
