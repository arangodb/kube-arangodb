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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	// UserNameRoot root user name
	UserNameRoot = "root"
)

// PasswordSecretName contains user password secret name
type PasswordSecretName string

func (p PasswordSecretName) Get() string {
	return string(p)
}

const (
	// PasswordSecretNameNone is magic value for no action
	PasswordSecretNameNone PasswordSecretName = "None"
	// PasswordSecretNameAuto is magic value for autogenerate name
	PasswordSecretNameAuto PasswordSecretName = "Auto"
)

// PasswordSecretNameList is a map from username to secretnames
type PasswordSecretNameList map[string]PasswordSecretName

// BootstrapSpec contains information for cluster bootstrapping
type BootstrapSpec struct {
	// PasswordSecretNames contains a map of username to password-secret-name
	PasswordSecretNames PasswordSecretNameList `json:"passwordSecretNames,omitempty"`
}

// IsNone returns true if p is None or p is empty
func (p PasswordSecretName) IsNone() bool {
	return p == PasswordSecretNameNone || p == ""
}

// IsAuto returns true if p is Auto
func (p PasswordSecretName) IsAuto() bool {
	return p == PasswordSecretNameAuto
}

// GetSecretName returns the secret name given by the specs. Or None if not set.
func (s PasswordSecretNameList) GetSecretName(user string) PasswordSecretName {
	if s != nil {
		if secretname, ok := s[user]; ok {
			return secretname
		}
	}
	return PasswordSecretNameNone
}

// getSecretNameForUserPassword returns the default secret name for the given user
func getSecretNameForUserPassword(deploymentname, username string) PasswordSecretName {
	return PasswordSecretName(shared.FixupResourceName(deploymentname + "-" + username + "-password"))
}

// Validate the specification.
func (b *BootstrapSpec) Validate() error {
	for username, secretname := range b.PasswordSecretNames {
		// Remove this restriction as soon as we can bootstrap databases
		if username != UserNameRoot {
			return errors.Newf("only username `root` allowed in passwordSecretNames")
		}

		if secretname.IsNone() {
			if username != UserNameRoot {
				return errors.Newf("magic value None not allowed for %s", username)
			}
		} else {
			if err := shared.ValidateResourceName(string(secretname)); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

// SetDefaults fills in default values when a field is not specified.
func (b *BootstrapSpec) SetDefaults(deploymentname string) {
	if b.PasswordSecretNames == nil {
		b.PasswordSecretNames = make(map[string]PasswordSecretName)
	}

	// If root is not set init with Auto
	if _, ok := b.PasswordSecretNames[UserNameRoot]; !ok {
		b.PasswordSecretNames[UserNameRoot] = PasswordSecretNameNone
	}

	// Replace Auto with generated secret name
	for user, secretname := range b.PasswordSecretNames {
		if secretname.IsAuto() {
			b.PasswordSecretNames[user] = getSecretNameForUserPassword(deploymentname, user)
		}
	}
}

// NewPasswordSecretNameListOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewPasswordSecretNameListOrNil(list PasswordSecretNameList) PasswordSecretNameList {
	if list == nil {
		return nil
	}
	var newList = make(PasswordSecretNameList)
	for k, v := range list {
		newList[k] = v
	}
	return newList
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (b *BootstrapSpec) SetDefaultsFrom(source BootstrapSpec) {
	if b.PasswordSecretNames == nil {
		b.PasswordSecretNames = NewPasswordSecretNameListOrNil(source.PasswordSecretNames)
	}
}
