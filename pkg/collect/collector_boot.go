//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package collect

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// eventTypeBoot is emitted once per pod boot.
const eventTypeBoot = "boot"

func init() {
	GetCollector().Register(bootCollector{})
}

// bootCollector emits a single event marking that the pod has booted. It is the canonical example
// of an ECollector and the minimal event the collector always pushes.
type bootCollector struct{}

// CollectEvents pushes a single boot event. The boot id dimension and the created timestamp are
// stamped centrally by the collector, so they are not set here.
func (bootCollector) CollectEvents(out util.Pusher[*Event]) error {
	out.Push(&Event{
		Type:      eventTypeBoot,
		ServiceId: serviceID,
	})

	return nil
}
