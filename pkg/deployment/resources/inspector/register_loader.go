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

package inspector

import (
	"sync"
)

var (
	inspectorLoadersList inspectorLoaders
	inspectorLoadersLock sync.Mutex
)

func requireRegisterInspectorLoader(i inspectorLoader) {
	if !registerInspectorLoader(i) {
		panic("Unable to register inspector loader")
	}
}

func registerInspectorLoader(i inspectorLoader) bool {
	inspectorLoadersLock.Lock()
	defer inspectorLoadersLock.Unlock()

	n := i.Name()

	if inspectorLoadersList.Get(n) != -1 {
		return false
	}

	inspectorLoadersList = append(inspectorLoadersList, i)

	return true
}

type inspectorLoaders []inspectorLoader
