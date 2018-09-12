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
// Author Ewout Prangsma
//

package resources

import (
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"
)

// Resources is a service that creates low level resources for members
// and inspects low level resources, put the inspection result in members.
type Resources struct {
	log     zerolog.Logger
	context Context
	health  struct {
		clusterHealth driver.ClusterHealth // Last fetched cluster health
		timestamp     time.Time            // Timestamp of last fetch of cluster health
		mutex         sync.Mutex           // Mutex guarding fields in this struct
	}
	enableFinalizers bool
}

// NewResources creates a new Resources service, used to
// create and inspect low level resources such as pods and services.
func NewResources(log zerolog.Logger, context Context, finalizers bool) *Resources {
	return &Resources{
		log:              log,
		context:          context,
		enableFinalizers: finalizers,
	}
}
