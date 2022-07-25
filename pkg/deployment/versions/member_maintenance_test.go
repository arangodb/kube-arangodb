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

package versions

import (
	"testing"
)

func Test_MemberMaintenance(t *testing.T) {
	runCheckTest(t, "MemberMaintenance - EE - 3.10", "3.10.0", true, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - CE - 3.10", "3.10.0", false, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - EE - 3.10.1", "3.10.1", true, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - CE - 3.10.1", "3.10.1", false, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - EE - 3.11", "3.11.0", true, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - CE - 3.11", "3.11.0", false, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - EE - 3.9", "3.9.0", true, false, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - CE - 3.9", "3.9.0", false, false, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - EE - 4.0", "4.0.0", true, true, memberMaintenanceChecker)
	runCheckTest(t, "MemberMaintenance - CE - 4.0", "4.0.0", false, true, memberMaintenanceChecker)
}
