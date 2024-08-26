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

package k8sutil

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	configMapsV1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func CreateConfigMap(ctx context.Context, configMaps configMapsV1.ModInterface, cm *core.ConfigMap, ownerRef *meta.OwnerReference) error {
	AddOwnerRefToObject(cm, ownerRef)

	if _, err := configMaps.Create(ctx, cm, meta.CreateOptions{}); err != nil {
		// Failed to create secret
		return kerrors.NewResourceError(err, cm)
	}
	return nil
}
