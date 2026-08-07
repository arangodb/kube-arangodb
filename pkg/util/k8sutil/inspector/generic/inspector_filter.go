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

package generic

import (
	"context"
	"sync"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/gvk"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type Inspector[S meta.Object] interface {
	gvk.GVK

	ListSimple() []S
	GetSimple(name string) (S, bool)
	Filter(filters ...Filter[S]) []S
	Iterate(action Action[S], filters ...Filter[S]) error
	Read() ReadClient[S]
}

// Subresource identifies a Kubernetes subresource targeted by a request. An empty value targets the
// main resource endpoint.
type Subresource string

// SubresourceStatus is the status subresource.
const SubresourceStatus Subresource = "status"

// String returns the subresource name.
func (s Subresource) String() string {
	return string(s)
}

// SubresourceInspector reports which subresources are exposed by the inspected resource.
type SubresourceInspector interface {
	// SubresourceEnabled reports whether the given subresource is exposed by the resource.
	SubresourceEnabled(subresource Subresource) bool
}

// SubresourceCacheTTL is how long the exposed subresource set is cached.
const SubresourceCacheTTL = 2 * time.Minute

// NewCachedSubresourceInspector returns a SubresourceInspector backed by fetch, caching the exposed
// subresource set for SubresourceCacheTTL. On a fetch error the last successfully fetched set is
// retained (so a transient failure never flips the reported state).
func NewCachedSubresourceInspector(fetch func(ctx context.Context) ([]Subresource, error)) SubresourceInspector {
	return &cachedSubresourceInspector{fetch: fetch}
}

type cachedSubresourceInspector struct {
	fetch func(ctx context.Context) ([]Subresource, error)

	lock sync.Mutex
	eol  time.Time
	set  map[Subresource]bool
}

func (i *cachedSubresourceInspector) SubresourceEnabled(subresource Subresource) bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	if time.Now().After(i.eol) {
		ctx, cancel := context.WithTimeout(shutdown.Context(), globals.DefaultKubernetesTimeout)
		subs, err := i.fetch(ctx)
		cancel()

		if err == nil {
			set := make(map[Subresource]bool, len(subs))
			for _, s := range subs {
				set[s] = true
			}

			i.set = set
			i.eol = time.Now().Add(SubresourceCacheTTL)
		}
		// On a fetch error keep the last-known set and retry on the next call.
	}

	return i.set[subresource]
}

type Filter[S meta.Object] func(obj S) bool
type Action[S meta.Object] func(obj S) error

func FilterObject[S meta.Object](obj S, filters ...Filter[S]) bool {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(obj) {
			return false
		}
	}

	return true
}

func FilterByLabels[S meta.Object](labels map[string]string) Filter[S] {
	return func(obj S) bool {
		objLabels := obj.GetLabels()
		for key, value := range labels {
			v, ok := objLabels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
