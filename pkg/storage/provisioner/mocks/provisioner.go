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

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Provisioner interface {
	provisioner.API
	MockGetter
}

type provisionerMock struct {
	mock.Mock
	nodeName            string
	available, capacity int64
	localPaths          map[string]struct{}
}

// NewProvisioner returns a new mocked provisioner
func NewProvisioner(nodeName string, available, capacity int64) Provisioner {
	return &provisionerMock{
		nodeName:   nodeName,
		available:  available,
		capacity:   capacity,
		localPaths: make(map[string]struct{}),
	}
}

func (m *provisionerMock) AsMock() *mock.Mock {
	return &m.Mock
}

// GetNodeInfo fetches information from the current node.
func (m *provisionerMock) GetNodeInfo(ctx context.Context) (provisioner.NodeInfo, error) {
	return provisioner.NodeInfo{
		NodeName: m.nodeName,
	}, nil
}

// GetInfo fetches information from the filesystem containing
// the given local path on the current node.
func (m *provisionerMock) GetInfo(ctx context.Context, localPath string) (provisioner.Info, error) {
	return provisioner.Info{
		NodeInfo: provisioner.NodeInfo{
			NodeName: m.nodeName,
		},
		Available: m.available,
		Capacity:  m.capacity,
	}, nil
}

// Prepare a volume at the given local path
func (m *provisionerMock) Prepare(ctx context.Context, localPath string) error {
	if _, found := m.localPaths[localPath]; found {
		return errors.Newf("Path already exists: %s", localPath)
	}
	m.localPaths[localPath] = struct{}{}
	return nil
}

// Remove a volume with the given local path
func (m *provisionerMock) Remove(ctx context.Context, localPath string) error {
	if _, found := m.localPaths[localPath]; !found {
		return errors.Newf("Path not found: %s", localPath)
	}
	delete(m.localPaths, localPath)
	return nil
}
