//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package v1alpha

import "fmt"

func (a *ArangoBackup) Validate() error {
	if err := a.Spec.Validate(); err != nil {
		return err
	}

	if err := a.Status.Validate(); err != nil {
		return err
	}

	return nil
}

func (a *ArangoBackupSpec) Validate() error {
	if a.Deployment.Name == "" {
		return fmt.Errorf("deployment name can not be empty")
	}

	if a.Download != nil {
		if err := a.Download.Validate(); err != nil {
			return err
		}
	}

	if a.Upload != nil {
		if err := a.Upload.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (a *ArangoBackupSpecOperation) Validate() error {
	if a.RepositoryPath == "" {
		return fmt.Errorf("RepositoryPath can not be empty")
	}

	return nil
}

func (a *ArangoBackupSpecDownload) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("ID can not be empty")
	}

	return a.ArangoBackupSpecOperation.Validate()
}

func (a *ArangoBackupStatus) Validate() error {
	if err := ArangoBackupStateMap.Exists(a.ArangoBackupState.State); err != nil {
		return err
	}

	return nil
}
