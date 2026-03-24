//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"bytes"
	"io"
	goStrings "strings"
	"text/template"

	"sigs.k8s.io/yaml"
)

type Template[T any] string

func (t Template[T]) RenderBytes(data T) ([]byte, error) {
	q := bytes.NewBuffer(nil)

	if err := t.Render(q, data); err != nil {
		return nil, err
	}

	return q.Bytes(), nil
}

func (t Template[T]) Render(out io.Writer, data T) error {
	z, err := template.New("config").Funcs(template.FuncMap{
		"toYaml": TemplateFuncToYaml,
		"indent": TemplateFuncIndent,
	}).Parse(string(t))
	if err != nil {
		return err
	}

	return z.Execute(out, data)
}

func TemplateFuncToYaml(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return goStrings.TrimRight(string(b), "\n"), nil
}

func TemplateFuncIndent(spaces int, s string) string {
	pad := goStrings.Repeat(" ", spaces)
	lines := goStrings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = pad + line
		}
	}
	return goStrings.Join(lines, "\n")
}
