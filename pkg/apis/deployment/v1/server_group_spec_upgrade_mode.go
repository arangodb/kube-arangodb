//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

// ServerGroupUpgradeMode is used to define Upgrade mode of the Pod
type ServerGroupUpgradeMode string

const (
	// ServerGroupUpgradeModeInplace define Inplace Upgrade procedure (with Upgrade initContainer).
	ServerGroupUpgradeModeInplace ServerGroupUpgradeMode = "inplace"

	// ServerGroupUpgradeModeReplace Replaces server instead of upgrading. Takes an effect only on DBServer
	ServerGroupUpgradeModeReplace ServerGroupUpgradeMode = "replace"

	// ServerGroupUpgradeModeOptionalReplace Replaces the member if upgrade fails with specific exit codes:
	// Code 30: In case of the Compaction Failure
	ServerGroupUpgradeModeOptionalReplace ServerGroupUpgradeMode = "optional-replace"

	// ServerGroupUpgradeModeManual Waits for the manual upgrade. Requires replacement or the annotation on the member.
	// Requires annotation `upgrade.deployment.arangodb.com/allow` on a Pod
	ServerGroupUpgradeModeManual ServerGroupUpgradeMode = "manual"

	// DefaultServerGroupUpgradeMode defaults to ServerGroupUpgradeModeInplace
	DefaultServerGroupUpgradeMode = ServerGroupUpgradeModeInplace
)

func (n *ServerGroupUpgradeMode) Validate() error {
	switch v := n.Get(); v {
	case ServerGroupUpgradeModeInplace, ServerGroupUpgradeModeReplace, ServerGroupUpgradeModeManual, ServerGroupUpgradeModeOptionalReplace:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown UpgradeMode %s", v.String()))
	}
}

func (n *ServerGroupUpgradeMode) Get() ServerGroupUpgradeMode {
	return n.Default(ServerGroupUpgradeModeInplace)
}

func (n *ServerGroupUpgradeMode) Default(d ServerGroupUpgradeMode) ServerGroupUpgradeMode {
	if n == nil {
		return d
	}

	return *n
}

func (n *ServerGroupUpgradeMode) String() string {
	return string(n.Get())
}

func (n *ServerGroupUpgradeMode) New() *ServerGroupUpgradeMode {
	v := n.Get()

	return &v
}
