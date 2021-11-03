//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestResources_SelectImage(t *testing.T) {
	type args struct {
		spec   v1.DeploymentSpec
		status v1.DeploymentStatus
		group  v1.ServerGroup
	}

	currentArangoDBImageInfo := v1.ImageInfo{
		Image:           "arango",
		ImageID:         "arangoID",
		ArangoDBVersion: "1.2.3",
		Enterprise:      true,
	}

	currentArangoSyncImageInfo := v1.ImageInfo{
		Image:           "arangosync",
		ImageID:         "arangosyncID",
		ArangoDBVersion: "2.6.0",
		Enterprise:      true,
	}

	tests := map[string]struct {
		args          args
		wantImageInfo v1.ImageInfo
		wantFound     bool
	}{
		"ArangoDB current image exists": {
			args: args{
				//spec: v1.DeploymentSpec{},
				status: v1.DeploymentStatus{
					CurrentImage:     &currentArangoDBImageInfo,
					CurrentSyncImage: &currentArangoSyncImageInfo,
					Images:           v1.ImageInfoList{currentArangoDBImageInfo},
					SyncImages:       v1.ImageInfoList{currentArangoDBImageInfo},
				},
				group: v1.ServerGroupAgents,
			},
			wantImageInfo: currentArangoDBImageInfo,
			wantFound:     true,
		},
		"ArangoSync current image exists": {
			args: args{
				spec: v1.DeploymentSpec{
					Sync: v1.SyncSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString("arangosync"),
					},
				},
				status: v1.DeploymentStatus{
					CurrentImage:     &currentArangoDBImageInfo,
					CurrentSyncImage: &currentArangoSyncImageInfo,
					Images:           v1.ImageInfoList{currentArangoDBImageInfo},
					SyncImages:       v1.ImageInfoList{currentArangoSyncImageInfo},
				},
				group: v1.ServerGroupSyncWorkers,
			},
			wantImageInfo: currentArangoSyncImageInfo,
			wantFound:     true,
		},
		"ArangoDB image not found": {
			args: args{
				spec: v1.DeploymentSpec{
					Sync: v1.SyncSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString("arangosync"),
					},
					Image: util.NewString("new"),
				},
				status: v1.DeploymentStatus{
					CurrentSyncImage: &currentArangoSyncImageInfo,
					Images:           v1.ImageInfoList{v1.ImageInfo{}},
					SyncImages:       v1.ImageInfoList{currentArangoSyncImageInfo},
				},
				group: v1.ServerGroupAgents,
			},
			wantImageInfo: v1.ImageInfo{},
			wantFound:     false,
		},
		"ArangoSync image not found": {
			args: args{
				spec: v1.DeploymentSpec{
					Sync: v1.SyncSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString("arangosync"),
					},
				},
				status: v1.DeploymentStatus{
					CurrentImage: &currentArangoDBImageInfo,
					Images:       v1.ImageInfoList{currentArangoDBImageInfo},
					SyncImages:   nil,
				},
				group: v1.ServerGroupSyncWorkers,
			},
			wantImageInfo: v1.ImageInfo{},
			wantFound:     false,
		},
		"ArangoDB image found": {
			args: args{
				spec: v1.DeploymentSpec{
					Sync: v1.SyncSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString("arangosync"),
					},
					Image: util.NewString("arango"),
				},
				status: v1.DeploymentStatus{
					Images:     v1.ImageInfoList{currentArangoDBImageInfo},
					SyncImages: v1.ImageInfoList{currentArangoSyncImageInfo},
				},
				group: v1.ServerGroupAgents,
			},
			wantImageInfo: currentArangoDBImageInfo,
			wantFound:     true,
		},
		"ArangoSync image found": {
			args: args{
				spec: v1.DeploymentSpec{
					Sync: v1.SyncSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString("arangosync"),
					},
					Image: util.NewString("arango"),
				},
				status: v1.DeploymentStatus{
					Images:     v1.ImageInfoList{currentArangoDBImageInfo},
					SyncImages: v1.ImageInfoList{currentArangoSyncImageInfo},
				},
				group: v1.ServerGroupSyncMasters,
			},
			wantImageInfo: currentArangoSyncImageInfo,
			wantFound:     true,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			r := &Resources{}
			imageInfo, found := r.SelectImage(tt.args.spec, tt.args.status, tt.args.group)
			require.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantImageInfo, imageInfo)
		})
	}
}
