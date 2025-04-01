//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package pretty

import (
	"encoding/json"
	"fmt"
	"reflect"
	goStrings "strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Table[T any] interface {
	json.Marshaler

	Add(in ...T) Table[T]

	RenderMarkdown() (string, error)
	Redner() (string, error)
}

type tableImpl[T any] struct {
	lock sync.Mutex

	rows []T
}

type tableJSONOut[T any] struct {
	Items []T `json:"items"`
}

func (t *tableImpl[T]) MarshalJSON() ([]byte, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	return json.Marshal(tableJSONOut[T]{Items: t.rows})
}

func (t *tableImpl[T]) Redner() (string, error) {
	o, err := t.table()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", o.Render()), nil
}

func (t *tableImpl[T]) RenderMarkdown() (string, error) {
	o, err := t.table()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", o.RenderMarkdown()), nil
}

func (t *tableImpl[T]) table() (table.Writer, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	z := reflect.TypeOf(util.Default[T]())

	if z.Kind() != reflect.Struct {
		return nil, errors.Errorf("Only Struct kind allowed")
	}

	var fields []int
	var columns []table.ColumnConfig

	for id := 0; id < z.NumField(); id++ {
		f := z.Field(id)

		if !f.IsExported() {
			continue
		}

		if f.Anonymous {
			continue
		}

		v, err := util.ExtractTags[TableTags](f)
		if err != nil {
			return nil, err
		}

		if v.Enabled == nil {
			return nil, nil
		}

		col, err := v.asColumnConfig()
		if err != nil {
			return nil, err
		}

		fields = append(fields, id)
		columns = append(columns, col)
	}

	wr := table.NewWriter()

	wr.AppendHeader(util.FormatList(columns, func(a table.ColumnConfig) interface{} {
		return a.Name
	}))

	wr.SetColumnConfigs(columns)

	wr.SuppressTrailingSpaces()

	for _, el := range t.rows {
		v := reflect.ValueOf(el)

		rows := make(table.Row, len(fields))

		for q, id := range fields {
			rows[q] = v.Field(id).Interface()
		}

		wr.AppendRow(rows)
	}

	return wr, nil
}

func (t *tableImpl[T]) Add(in ...T) Table[T] {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.rows = append(t.rows, in...)

	return t
}

type TableTags struct {
	Enabled     *string `tag:"table"`
	Align       *string `tag:"table_align"`
	HeaderAlign *string `tag:"table_header_align"`
}

func (t TableTags) asColumnConfig() (table.ColumnConfig, error) {
	var r table.ColumnConfig

	r.Name = *t.Enabled

	if a := t.Align; a != nil {
		switch v := goStrings.ToLower(*a); v {
		case "left":
			r.Align = text.AlignLeft
		case "right":
			r.Align = text.AlignRight
		case "center":
			r.Align = text.AlignCenter
		default:
			return table.ColumnConfig{}, errors.Errorf("Unsuported align format: %s", v)
		}
	}

	if a := t.HeaderAlign; a != nil {
		switch v := goStrings.ToLower(*a); v {
		case "left":
			r.AlignHeader = text.AlignLeft
		case "right":
			r.AlignHeader = text.AlignRight
		case "center":
			r.AlignHeader = text.AlignCenter
		default:
			return table.ColumnConfig{}, errors.Errorf("Unsuported align format: %s", v)
		}
	}

	return r, nil
}

func NewTable[T any]() Table[T] {
	return &tableImpl[T]{}
}
