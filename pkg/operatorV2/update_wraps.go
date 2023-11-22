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

package operator

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
)

func WithArangoBackupUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup], obj *backupApi.ArangoBackup, status backupApi.ArangoBackupStatus, opts meta.UpdateOptions) (*backupApi.ArangoBackup, error) {
	return WithUpdateStatusInterfaceRetry[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup](ctx, client, obj, status, opts)
}

func WithArangoExtensionUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[mlApi.ArangoMLExtensionStatus, *mlApi.ArangoMLExtension], obj *mlApi.ArangoMLExtension, status mlApi.ArangoMLExtensionStatus, opts meta.UpdateOptions) (*mlApi.ArangoMLExtension, error) {
	return WithUpdateStatusInterfaceRetry[mlApi.ArangoMLExtensionStatus, *mlApi.ArangoMLExtension](ctx, client, obj, status, opts)
}
