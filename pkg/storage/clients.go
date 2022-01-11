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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createProvisionerClients creates a list of clients for all known
// provisioners.
func (ls *LocalStorage) createProvisionerClients() ([]provisioner.API, error) {
	// Find provisioner endpoints
	ns := ls.apiObject.GetNamespace()
	listOptions := k8sutil.LocalStorageListOpt(ls.apiObject.GetName(), roleProvisioner)
	items, err := ls.deps.KubeCli.CoreV1().Endpoints(ns).List(context.Background(), listOptions)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	addrs := createValidEndpointList(items)
	if len(addrs) == 0 {
		// No provisioners available
		return nil, nil
	}
	// Create clients for endpoints
	clients := make([]provisioner.API, len(addrs))
	for i, addr := range addrs {
		var err error
		clients[i], err = client.New(fmt.Sprintf("http://%s", addr))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return clients, nil
}

// GetClientByNodeName looks for a client that serves the given node name.
// Returns an error if no such client is found.
func (ls *LocalStorage) GetClientByNodeName(nodeName string) (provisioner.API, error) {
	clients, err := ls.createProvisionerClients()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Find matching client
	for _, c := range clients {
		ctx := context.Background()
		if info, err := c.GetNodeInfo(ctx); err == nil && info.NodeName == nodeName {
			return c, nil
		}
	}
	return nil, errors.WithStack(errors.Newf("No client found for node name '%s'", nodeName))
}
