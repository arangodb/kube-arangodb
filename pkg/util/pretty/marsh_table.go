//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"reflect"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Table[T any] interface {
	Add(in ...T) Table[T]

	RenderMarkdown() string
}

type tableImpl[T any] struct {
	wr table.Writer

	fields []int
}

func (t tableImpl[T]) RenderMarkdown() string {
	return fmt.Sprintf("%s\n", t.wr.RenderMarkdown())
}

func (t tableImpl[T]) Add(in ...T) Table[T] {
	for _, el := range in {
		v := reflect.ValueOf(el)

		rows := make(table.Row, len(t.fields))

		for q, id := range t.fields {
			rows[q] = v.Field(id).Interface()
		}

		t.wr.AppendRow(rows)
	}

	return t
}

type TableTags struct {
	Enabled *string `tag:"table"`
	Align   *string `tag:"table_align"`
}

func (t TableTags) asColumnConfig() (table.ColumnConfig, error) {
	var r table.ColumnConfig

	r.Name = *t.Enabled

	if a := t.Align; a != nil {
		switch v := strings.ToLower(*a); v {
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

	return r, nil
}

func NewTable[T any]() (Table[T], error) {
	t := reflect.TypeOf(util.Default[T]())

	if t.Kind() != reflect.Struct {
		return nil, errors.Errorf("Only Struct kind allowed")
	}

	var fields []int
	var columns []table.ColumnConfig

	for id := 0; id < t.NumField(); id++ {
		f := t.Field(id)

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

	return tableImpl[T]{
		wr:     wr,
		fields: fields,
	}, nil
}
