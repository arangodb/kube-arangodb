//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package arangoplatformstorage

import (
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/base"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

type Inspector interface {
	ArangoPlatformStorage() Definition
}

type Definition interface {
	base.Inspector

	V1Beta1() (generic.Inspector[*platformApi.ArangoPlatformStorage], error)
}
