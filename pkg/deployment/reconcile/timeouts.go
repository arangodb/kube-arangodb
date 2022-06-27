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

package reconcile

import "time"

const (
	defaultTimeout = time.Minute * 10

	addMemberTimeout                 = time.Minute * 10
	backupRestoreTimeout             = time.Minute * 15
	cleanoutMemberTimeout            = time.Hour * 12
	operationTLSCACertificateTimeout = time.Minute * 30
	pingTimeout                      = time.Minute * 15
	pvcResizeTimeout                 = time.Minute * 30
	pvcResizedTimeout                = time.Minute * 15
	recreateMemberTimeout            = time.Minute * 15
	removeMemberTimeout              = time.Minute * 15
	rotateMemberTimeout              = time.Minute * 15
	shutdownMemberTimeout            = time.Minute * 30
	shutdownTimeout                  = time.Second * 15
	tlsSNIUpdateTimeout              = time.Minute * 10
	upgradeMemberTimeout             = time.Hour * 6
	waitForMemberUpTimeout           = time.Minute * 30
)
