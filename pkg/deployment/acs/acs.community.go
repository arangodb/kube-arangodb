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
//go:build !enterprise
// +build !enterprise

package acs

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewACS(main types.UID, cache inspectorInterface.Inspector) sutil.ACS {
	return acs{
		main:  main,
		cache: cache,
	}
}

type acs struct {
	main  types.UID
	cache inspectorInterface.Inspector
}

func (a acs) ForEachHealthyCluster(f func(item sutil.ACSItem) error) error {
	return f(a)
}

func (a acs) CurrentClusterCache() inspectorInterface.Inspector {
	return a.cache
}

func (a acs) ClusterCache(uid types.UID) (inspectorInterface.Inspector, bool) {
	c, ok := a.Cluster(uid)
	if ok {
		return c.Cache(), true
	}

	return nil, false
}

func (a acs) UID() types.UID {
	return a.main
}

func (a acs) Ready() bool {
	return true
}

func (a acs) Cache() inspectorInterface.Inspector {
	return a.cache
}

func (a acs) Cluster(uid types.UID) (sutil.ACSItem, bool) {
	if a.main == uid || uid == "" {
		return a, true
	}

	return nil, false
}

func (a acs) RemoteClusters() []types.UID {
	return nil
}

func (a acs) Inspect(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector) error {
	return nil
}
