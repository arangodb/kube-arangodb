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

package resources

import (
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
	clientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
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
	shardSync struct {
		allInSync             bool
		timestamp             time.Time
		mutex                 sync.Mutex
		triggerSyncInspection trigger.Trigger
	}
	monitoringClient *clientv1.MonitoringV1Client
}

// NewResources creates a new Resources service, used to
// create and inspect low level resources such as pods and services.
func NewResources(log zerolog.Logger, context Context) *Resources {
	return &Resources{
		log:     log,
		context: context,
	}
}
