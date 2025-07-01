//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package backup

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func newEventInstance(recorder event.Recorder) event.RecorderInstance {
	return recorder.NewInstance(backupApi.SchemeGroupVersion.Group,
		backupApi.SchemeGroupVersion.Version,
		backup.ArangoBackupResourceKind)
}

// RegisterInformer into operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory) error {
	if err := operator.RegisterInformer(informer.Backup().V1().ArangoBackups().Informer(),
		backupApi.SchemeGroupVersion.Group,
		backupApi.SchemeGroupVersion.Version,
		backup.ArangoBackupResourceKind); err != nil {
		return err
	}

	h := &handler{
		client:     client.Arango(),
		kubeClient: client.Kubernetes(),

		eventRecorder: newEventInstance(recorder),

		operator: operator,
	}
	h.arangoClientFactory = newArangoClientBackupFactory(h)

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	if err := operator.RegisterStarter(h); err != nil {
		return err
	}

	return nil
}
