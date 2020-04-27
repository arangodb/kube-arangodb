//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package reconcile

import "time"

const (
	addMemberTimeout             = time.Minute * 5
	cleanoutMemberTimeout        = time.Hour * 12
	removeMemberTimeout          = time.Minute * 15
	recreateMemberTimeout        = time.Minute * 15
	renewTLSCertificateTimeout   = time.Minute * 30
	renewTLSCACertificateTimeout = time.Minute * 30
	rotateMemberTimeout          = time.Minute * 15
	pvcResizeTimeout             = time.Minute * 15
	pvcResizedTimeout            = time.Minute * 15
	shutdownMemberTimeout        = time.Minute * 30
	upgradeMemberTimeout         = time.Hour * 6
	waitForMemberUpTimeout       = time.Minute * 15
	tlsSNIUpdateTimeout          = time.Minute * 10

	shutdownTimeout = time.Second * 15
)
