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

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner/client"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type Clients map[string]provisioner.API

func (c Clients) Copy() Clients {
	r := make(Clients, len(c))

	for k, v := range c {
		r[k] = v
	}

	return r
}

func (c Clients) Filter(f func(node string, client provisioner.API) bool) Clients {
	r := make(Clients, len(c))

	for k, v := range c {
		if f(k, v) {
			r[k] = v
		}
	}

	return r
}

func (c Clients) Keys() []string {
	r := make([]string, 0, len(c))

	for k := range c {
		r = append(r, k)
	}

	return r
}

// createProvisionerClients creates a list of clients for all known
// provisioners.
func (ls *LocalStorage) createProvisionerClients(ctx context.Context) (Clients, error) {
	// Find provisioner endpoints
	ns := ls.apiObject.GetNamespace()
	listOptions := k8sutil.LocalStorageListOpt(ls.apiObject.GetName(), roleProvisioner)
	items, err := ls.deps.Client.Kubernetes().CoreV1().Endpoints(ns).List(context.Background(), listOptions)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	addrs := createValidEndpointList(items)
	if len(addrs) == 0 {
		// No provisioners available
		return nil, nil
	}
	// Create clients for endpoints
	clients := make(map[string]provisioner.API, len(addrs))
	for _, addr := range addrs {
		c, err := client.New(fmt.Sprintf("http://%s", addr))
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if info, err := ls.fetchClientNodeInfo(ctx, c); err == nil {
			clients[info.NodeName] = c
		}
	}
	return clients, nil
}

func (ls *LocalStorage) fetchClientNodeInfo(ctx context.Context, c provisioner.API) (provisioner.NodeInfo, error) {
	nctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.GetNodeInfo(nctx)
}

// GetClientByNodeName looks for a client that serves the given node name.
// Returns an error if no such client is found.
func (ls *LocalStorage) GetClientByNodeName(ctx context.Context, nodeName string) (provisioner.API, error) {
	clients, err := ls.createProvisionerClients(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Find matching client
	if c, ok := clients[nodeName]; ok {
		return c, nil
	}

	return nil, errors.WithStack(errors.Newf("No client found for node name '%s'", nodeName))
}
