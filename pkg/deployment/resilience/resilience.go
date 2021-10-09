//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package resilience

import "github.com/rs/zerolog"

// Resilience is the service that inspects the overall state of the deployment
// to improve resilience.
type Resilience struct {
	log     zerolog.Logger
	context Context
}

// NewResilience creates a new resilience with given context.
func NewResilience(log zerolog.Logger, context Context) *Resilience {
	return &Resilience{
		log:     log,
		context: context,
	}
}
