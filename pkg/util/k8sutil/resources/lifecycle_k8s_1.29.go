//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

//go:build !kube_arangodb_k8s_1_28

package resources

import core "k8s.io/api/core/v1"

func MergeLifecycleHandler(a, b *core.LifecycleHandler) *core.LifecycleHandler {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b.DeepCopy()
	}
	if b == nil {
		return a.DeepCopy()
	}

	if a.HTTPGet != nil || a.Exec != nil || a.TCPSocket != nil || a.Sleep != nil {
		return a.DeepCopy()
	}

	return b.DeepCopy()
}
