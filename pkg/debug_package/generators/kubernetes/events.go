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

package kubernetes

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	eventsv1 "k8s.io/api/events/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Events() shared.Factory {
	return shared.NewFactory("kubernetes-events", func(cmd *cobra.Command) {
		f := cmd.Flags()
		if f.Lookup("namespace") == nil {
			f.String("namespace", "default", "Kubernetes namespace")
		}
	}, func(cmd *cobra.Command, logger zerolog.Logger, files chan<- shared.File) error {
		k, ok := kclient.GetDefaultFactory().Client()
		if !ok {
			return errors.Newf("Client is not initialised")
		}

		ns, _ := cmd.Flags().GetString("namespace")

		events := map[types.UID]*eventsv1.Event{}
		next := ""
		for {
			r, err := k.Kubernetes().EventsV1().Events(ns).List(context.Background(), meta.ListOptions{
				Continue: next,
			})

			if err != nil {
				return err
			}

			for _, e := range r.Items {
				events[e.UID] = e.DeepCopy()
			}

			next = r.Continue
			if next == "" {
				break
			}
		}

		files <- shared.NewJSONFile("kubernetes/events.json", func() (interface{}, error) {
			q := make([]*eventsv1.Event, 0, len(events))

			for _, e := range events {
				q = append(q, e)
			}

			return q, nil
		})

		return nil
	})
}
