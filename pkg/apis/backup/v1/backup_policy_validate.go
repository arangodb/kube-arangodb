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

package v1

import (
	"time"

	"github.com/robfig/cron"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (a *ArangoBackupPolicy) Validate() error {
	if err := a.Spec.Validate(); err != nil {
		return err
	}

	return nil
}

func (a *ArangoBackupPolicySpec) Validate() error {
	if expr, err := cron.ParseStandard(a.Schedule); err != nil {
		return errors.Newf("error while parsing expr: %s", err.Error())
	} else if expr.Next(time.Now()).IsZero() {
		return errors.Newf("invalid schedule format")
	}

	return nil
}
