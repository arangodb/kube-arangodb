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

package deployment

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/names"
)

func (d *Deployment) createInitialTopology(ctx context.Context) error {
	return nil
}

func (d *Deployment) renderMemberID(_ api.DeploymentSpec, status *api.DeploymentStatus, _ *api.ServerGroupStatus, group api.ServerGroup) string {
	for {
		if id := names.GetArangodID(group); !status.Members.ContainsID(id) {
			return id
		}
	}
}
