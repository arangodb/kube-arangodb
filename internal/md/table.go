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

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewTable(columns ...Column) Table {
	return &table{
		columns: columns,
	}
}

type Table interface {
	Render() string

	AddRow(in map[Column]string) error
}

type table struct {
	lock sync.Mutex

	columns Columns
	rows    []map[string]string
}

func (t *table) AddRow(in map[Column]string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	r := map[string]string{}

	for k, v := range in {
		if _, ok := t.columns.Get(k.ID()); !ok {
			return errors.Newf("Column not found")
		}

		r[k.ID()] = v
	}

	t.rows = append(t.rows, r)

	return nil
}

func (t *table) fillString(base, filler string, align ColumnAlign, size int) string {
	for len(base) < size {
		switch align {
		case ColumnLeftAlign:
			base += filler
		case ColumnRightAlign:
			base = filler + base
		case ColumnCenterAlign:
			base += filler
			if len(base) < size {
				base = filler + base
			}
		}
	}

	return base
}

func (t *table) Render() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	ks := map[string]int{}

	for _, c := range t.columns {
		ks[c.ID()] = len(c.Name())
	}

	for _, r := range t.rows {
		for _, c := range t.columns {
			if q := len(r[c.ID()]); q > ks[c.ID()] {
				ks[c.ID()] = q
			}
		}
	}

	buff := ""

	buff += "|"

	for _, c := range t.columns {
		buff += " "
		buff += t.fillString(c.Name(), " ", c.Align(), ks[c.ID()])
		buff += " |"
	}
	buff += "\n|"

	for _, c := range t.columns {
		switch c.Align() {
		case ColumnLeftAlign, ColumnCenterAlign:
			buff += ":"
		default:
			buff += "-"
		}

		buff += t.fillString("", "-", ColumnLeftAlign, ks[c.ID()])
		switch c.Align() {
		case ColumnRightAlign, ColumnCenterAlign:
			buff += ":"
		default:
			buff += "-"
		}
		buff += "|"
	}
	buff += "\n"

	for _, r := range t.rows {
		buff += "|"

		for _, c := range t.columns {
			buff += " "
			buff += t.fillString(r[c.ID()], " ", c.Align(), ks[c.ID()])
			buff += " |"
		}
		buff += "\n"
	}

	return buff
}
