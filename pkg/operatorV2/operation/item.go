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

package operation

import (
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// Operation declares operation string representation
type Operation string

const (
	// Add define operation generated when object was created
	Add Operation = "ADD"
	// Update define operation generated when object was updated
	Update Operation = "UPDATE"
	// Delete define operation generated when object was deleted
	Delete Operation = "DELETE"

	emptyError            = "value %s cannot be empty"
	invalidCharacterError = "character %s is not allowed into %s"

	separator = "/"
)

// NewItemFromString creates new item from String
func NewItemFromString(itemString string) (Item, error) {
	parts := strings.Split(itemString, "/")

	if len(parts) != 6 {
		return Item{}, errors.Newf("expected 6 parts in %s, got %d", itemString, len(parts))
	}

	return NewItem(Operation(parts[0]), parts[1], parts[2], parts[3], parts[4], parts[5])
}

// NewItemFromObject creates new item from Kubernetes Object
func NewItemFromObject(operation Operation, group, version, kind string, object meta.Object) (Item, error) {
	return NewItem(operation, group, version, kind, object.GetNamespace(), object.GetName())
}

// NewItem creates new Item
func NewItem(operation Operation, group, version, kind, namespace, name string) (Item, error) {
	i := Item{
		Operation: operation,
		Group:     group,
		Version:   version,
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
	}

	if err := i.Validate(); err != nil {
		return Item{}, err
	}

	return i, nil
}

// Item defines action in operator
type Item struct {
	Operation Operation

	Group   string
	Version string
	Kind    string

	Namespace string
	Name      string
}

func validateField(name, value string, allowEmpty bool) error {
	if !allowEmpty && value == "" {
		return errors.Newf(emptyError, name)
	}

	if index := strings.Index(value, separator); index != -1 {
		return errors.Newf(invalidCharacterError, separator, name)
	}

	return nil
}

// Validate item if all required fields are set
func (i Item) Validate() error {
	if err := validateField("operation", string(i.Operation), false); err != nil {
		return err
	}

	if err := validateField("group", i.Group, true); err != nil {
		return err
	}
	if err := validateField("version", i.Version, false); err != nil {
		return err
	}
	if err := validateField("kind", i.Kind, false); err != nil {
		return err
	}

	if err := validateField("namespace", i.Namespace, true); err != nil {
		return err
	}
	if err := validateField("name", i.Name, false); err != nil {
		return err
	}

	return nil
}

func (i Item) String() string {
	return strings.Join([]string{string(i.Operation), i.Group, i.Version, i.Kind, i.Namespace, i.Name}, separator)
}
