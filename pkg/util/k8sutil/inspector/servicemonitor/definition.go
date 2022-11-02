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

package servicemonitor

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/anonymous"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/gvk"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/refresh"
	v1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

type Inspector interface {
	ServiceMonitor() Definition
}

type Definition interface {
	refresh.Inspector

	gvk.GK
	anonymous.Impl

	Version() version.Version

	V1() (v1.Inspector, error)
}
