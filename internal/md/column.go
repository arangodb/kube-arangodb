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

package md

import "k8s.io/apimachinery/pkg/util/uuid"

type ColumnAlign int

const (
	ColumnRightAlign ColumnAlign = iota
	ColumnCenterAlign
	ColumnLeftAlign
)

type Columns []Column

func (c Columns) Get(id string) (Column, bool) {
	for _, z := range c {
		if z.ID() == id {
			return z, true
		}
	}

	return nil, false
}

type Column interface {
	Name() string
	Align() ColumnAlign

	ID() string
}

func NewColumn(name string, align ColumnAlign) Column {
	return column{
		name:  name,
		id:    string(uuid.NewUUID()),
		align: align,
	}
}

type column struct {
	name  string
	id    string
	align ColumnAlign
}

func (c column) ID() string {
	return c.id
}

func (c column) Name() string {
	return c.name
}

func (c column) Align() ColumnAlign {
	return c.align
}
