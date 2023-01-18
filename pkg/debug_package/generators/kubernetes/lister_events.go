//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type EventList map[types.UID]*core.Event

func (d EventList) AsList() []*core.Event {
	eventList := make([]*core.Event, 0, len(d))
	for _, p := range d {
		eventList = append(eventList, p)
	}

	return eventList
}

func ListEvents(k kclient.Client) (EventList, error) {
	events := EventList{}
	next := ""

	for {
		deps, err := k.Kubernetes().CoreV1().Events(cli.GetInput().Namespace).List(context.Background(), meta.ListOptions{
			Continue: next,
		})
		if err != nil {
			return nil, err
		}

		for _, d := range deps.Items {
			events[d.UID] = d.DeepCopy()
			events[d.UID].ManagedFields = nil
		}

		if deps.Continue == "" {
			break
		}

		next = deps.Continue
	}

	return events, nil
}
