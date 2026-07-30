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
	"context"
	"crypto/tls"
	goHttp "net/http"
	"time"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"

	pbImplEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1"
	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	arangodClient "github.com/arangodb/kube-arangodb/pkg/util/arangod/client"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	// eventsCollectionName is the system collection the startup event is stored in. It matches the
	// collection the events integration uses, so both write in the same place and format.
	eventsCollectionName = "_events"

	// eventsTTLIndexName matches the TTL index created by the events integration.
	eventsTTLIndexName = "system_events_created_ttl_index"
)

// emit connects directly to the ArangoDB endpoint, ensures the _events system collection exists
// (creating it, and its created-TTL index, if missing) and inserts the startup event. It reuses the
// events integration collection builder and remote store so the collection layout and stored
// document format are identical whether the collector or the integration wrote the event.
//
// A failure is returned so the collector run loop retries on the next cycle - during early pod
// startup arangod may not be serving yet.
func emit(ctx context.Context, opts Options, event *pbEventsV1.Event) error {
	factory := arangodClient.NewFactory(
		arangodClient.FolderArangoDBAuthentication(opts.JWTPath),
		arangodClient.HTTPClientFactory(func(t *goHttp.Transport) {
			// The local arangod endpoint serves a self-signed certificate when TLS is enabled.
			t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}),
	)

	c, err := factory.Client(ctx, opts.Endpoint)
	if err != nil {
		return errors.Wrapf(err, "unable to connect to arangodb at %s", opts.Endpoint)
	}

	clientObj := cache.NewObject[adbDriverV2.Client](func(ctx context.Context) (adbDriverV2.Client, time.Duration, error) {
		return c, time.Hour, nil
	})

	col := db.NewClient(clientObj).Database("_system").
		CreateCollection(eventsCollectionName, db.StaticProps(adbDriverV2.CreateCollectionPropertiesV2{IsSystem: util.NewType(true)})).
		WithTTLIndex(eventsTTLIndexName, pbImplEventsV1.DefaultTTL, "created").
		Get()

	store := pbImplEventsV1.NewArangoRemoteStore[*pbEventsV1.Event](col)

	if err := store.Init(ctx); err != nil {
		return errors.Wrapf(err, "unable to ensure %s collection", eventsCollectionName)
	}

	if err := store.Emit(ctx, event); err != nil {
		return errors.Wrapf(err, "unable to insert startup event")
	}

	return nil
}
