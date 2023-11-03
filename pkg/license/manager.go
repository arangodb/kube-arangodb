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

package license

import (
	"context"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

var (
	lock            sync.Mutex
	managerInstance Manager
)

func initManager(config Config, loader Loader) error {
	lock.Lock()
	defer lock.Unlock()

	if managerInstance != nil {
		return errors.Newf("Manager is already initialised")
	}

	mgr := &manager{
		loader:  loader,
		config:  config,
		license: StatusMissing,
	}

	go mgr.run(shutdown.Channel())

	managerInstance = mgr

	return nil
}

func ManagerInstance() Manager {
	lock.Lock()
	defer lock.Unlock()

	if managerInstance == nil {
		panic("Manager not yet initialised")
	}

	return managerInstance
}

type Manager interface {
}

type manager struct {
	config Config
	loader Loader

	license License
}

func (m *manager) run(stopC <-chan struct{}) {
	t := time.NewTicker(m.config.RefreshInterval)
	defer t.Stop()

	for {
		select {
		case <-stopC:
			return
		case <-t.C:
			// Lets do the refresh
			logger.Debug("Refresh process started")

			license, ok, err := m.loadLicense()
			if err != nil {
				logger.Err(err).Warn("Unable to load license")
				continue
			}

			if !ok {
				logger.Debug("License is missing")
				continue
			}

			m.license = checkLicense(license)
		}
	}
}

func (m *manager) loadLicense() (string, bool, error) {
	ctx, c := context.WithTimeout(context.Background(), m.config.RefreshTimeout)
	defer c()
	return m.loader.Refresh(ctx)
}
