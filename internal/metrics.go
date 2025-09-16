//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package internal

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

//go:embed metrics.go.tmpl
var metricsGoTemplate []byte

//go:embed metrics.item.go.tmpl
var metricsItemGoTemplate []byte

//go:embed metrics.item.go_test.tmpl
var metricsItemGoTestTemplate []byte

//go:embed metrics.item.tmpl
var metricItemTemplate []byte

//go:embed metrics.yaml
var metricsData []byte

type MetricsDoc struct {
	Destination   string `json:"destination" yaml:"destination"`
	Documentation string `json:"documentation" yaml:"documentation"`

	Namespaces Namespaces `json:"namespaces" yaml:"namespaces"`
}

type Namespaces map[string]Groups

func (n Namespaces) Keys() []string {
	r := make([]string, 0, len(n))

	for k := range n {
		r = append(r, k)
	}

	sort.Strings(r)

	return r
}

type Groups map[string]Metrics

func (n Groups) Keys() []string {
	r := make([]string, 0, len(n))

	for k := range n {
		r = append(r, k)
	}

	sort.Strings(r)

	return r
}

type Metrics map[string]Metric

func (n Metrics) Keys() []string {
	r := make([]string, 0, len(n))

	for k := range n {
		r = append(r, k)
	}

	sort.Strings(r)

	return r
}

type Metric struct {
	Description      string `json:"description" yaml:"description"`
	Type             string `json:"type" yaml:"type"`
	ShortDescription string `json:"shortDescription" yaml:"shortDescription"`

	Global bool `json:"global" yaml:"global"`

	Labels        []Label    `json:"labels" yaml:"labels"`
	AlertingRules []Alerting `json:"alertingRules" yaml:"alertingRules"`
}

type Alerting struct {
	Priority    string `json:"priority" yaml:"priority"`
	Query       string `json:"query" yaml:"query"`
	Description string `json:"description" yaml:"description"`
}

type Label struct {
	Key         string   `json:"key" yaml:"key"`
	Description string   `json:"description" yaml:"description"`
	Type        *string  `json:"type" yaml:"type"`
	Values      []string `json:"values" yaml:"values"`
}

func GenerateMetricsDocumentation(root string, in MetricsDoc) error {
	docsRoot := path.Join(root, in.Documentation)
	goFilesRoot := path.Join(root, in.Destination)

	if _, err := os.Stat(docsRoot); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(docsRoot, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if _, err := os.Stat(goFilesRoot); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(goFilesRoot, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := generateMetricsREADME(docsRoot, in); err != nil {
		return err
	}

	if err := generateMetricsGO(goFilesRoot, in); err != nil {
		return err
	}

	return nil
}

func generateMetricFile(root, name string, m Metric) error {
	type tableRowLabels struct {
		Label       string `table:"Label" table_align:"center"`
		Description string `table:"Description" table_align:"left"`
		Values      string `table:"Values" table_align:"center"`
	}
	type tableRowPriority struct {
		Priority    string `table:"Priority" table_align:"center"`
		Query       string `table:"Query" table_align:"center"`
		Description string `table:"Description" table_align:"left"`
	}

	t := pretty.NewTable[tableRowLabels]()

	ta := pretty.NewTable[tableRowPriority]()

	for _, l := range m.Labels {
		t.Add(tableRowLabels{
			Label:       l.Key,
			Description: l.Description,
			Values:      util.BoolSwitch(len(l.Values) == 0, "*", strings.Join(l.Values, "\n")),
		})
	}

	for _, l := range m.AlertingRules {
		//nolint:all
		ta.Add(tableRowPriority{
			Priority:    l.Priority,
			Query:       l.Query,
			Description: l.Description,
		})
	}

	q, err := template.New("metrics").Parse(string(metricItemTemplate))
	if err != nil {
		return err
	}

	out, err := os.OpenFile(path.Join(root, fmt.Sprintf("%s.md", name)), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	tr, err := t.RenderMarkdown()
	if err != nil {
		return err
	}

	tar, err := ta.RenderMarkdown()
	if err != nil {
		return err
	}

	if err := q.Execute(out, map[string]interface{}{
		"name":           name,
		"type":           m.Type,
		"description":    m.Description,
		"labels_table":   tr,
		"labels":         len(m.Labels) > 0,
		"alerting_table": tar,
		"alerting":       len(m.AlertingRules) > 0,
	}); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

func generateMetricsREADME(root string, in MetricsDoc) error {
	type tableRow struct {
		Name        string `table:"Name" table_align:"center"`
		Namespace   string `table:"Namespace" table_align:"center"`
		Group       string `table:"Group" table_align:"center"`
		Type        string `table:"Type" table_align:"center"`
		Description string `table:"Description" table_align:"left"`
	}

	t := pretty.NewTable[tableRow]()

	for _, namespace := range in.Namespaces.Keys() {
		for _, g := range in.Namespaces[namespace].Keys() {
			for _, metric := range in.Namespaces[namespace][g].Keys() {
				mname := fmt.Sprintf("%s_%s_%s", namespace, g, metric)
				rname := fmt.Sprintf("[%s](./%s.md)", mname, mname)

				details := in.Namespaces[namespace][g][metric]

				t.Add(tableRow{
					Name:        rname,
					Namespace:   namespace,
					Group:       g,
					Type:        details.Type,
					Description: details.ShortDescription,
				})

				if err := generateMetricFile(root, mname, details); err != nil {
					return err
				}
			}
		}
	}

	o, err := t.RenderMarkdown()
	if err != nil {
		return err
	}

	if err := pretty.ReplaceSectionsInFile(path.Join(root, "README.md"), map[string]string{
		"metricsTable": pretty.WrapWithNewLines(o),
	}); err != nil {
		return err
	}

	return nil
}

func generateLabels(labels []Label) string {
	if len(labels) == 0 {
		return "nil"
	}

	parts := make([]string, len(labels))

	for id := range labels {
		parts[id] = fmt.Sprintf("`%s`", labels[id].Key)
	}

	return fmt.Sprintf("[]string{%s}", strings.Join(parts, ", "))
}

func generateMetricsGO(root string, in MetricsDoc) error {
	i, err := template.New("metrics").Parse(string(metricsItemGoTemplate))
	if err != nil {
		return err
	}

	t, err := template.New("metrics").Parse(string(metricsItemGoTestTemplate))
	if err != nil {
		return err
	}

	for _, namespace := range in.Namespaces.Keys() {
		for _, g := range in.Namespaces[namespace].Keys() {
			for _, metric := range in.Namespaces[namespace][g].Keys() {
				details := in.Namespaces[namespace][g][metric]

				mname := fmt.Sprintf("%s_%s_%s", namespace, g, metric)

				parts := strings.Split(mname, "_")
				tparts := strings.Split(strings.Title(strings.Join(parts, " ")), " ")

				fnameParts := make([]string, len(parts))
				for id := range parts {
					if id == 0 {
						fnameParts[id] = parts[id]
					} else {
						fnameParts[id] = tparts[id]
					}
				}

				var keys []string
				var params []string

				params = append(params, "value float64")
				keys = append(keys, "value")

				var mapTypes = map[string]string{}
				var mapKeys []string
				var mapIKeys = map[string]string{}

				for _, label := range details.Labels {
					v := strings.Split(strings.ToLower(label.Key), "_")
					for id := range v {
						if id == 0 {
							continue
						}

						v[id] = strings.Title(v[id])
					}

					k := strings.Join(v, "")

					v[0] = strings.Title(v[0])

					kPublic := strings.Join(v, "")

					keys = append(keys, k)
					mapKeys = append(mapKeys, kPublic)

					mapIKeys[kPublic] = k

					if t := label.Type; t != nil {
						params = append(params, fmt.Sprintf("%s %s", k, *t))
						mapTypes[kPublic] = *t
					} else {
						params = append(params, fmt.Sprintf("%s string", k))
						mapTypes[kPublic] = "string"
					}
				}

				save := func(template *template.Template, path string) error {
					out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
					if err != nil {
						return err
					}

					if err := template.Execute(out, map[string]interface{}{
						"name":             mname,
						"fname":            strings.Join(fnameParts, ""),
						"ename":            strings.Join(tparts, ""),
						"shortDescription": details.ShortDescription,
						"labels":           generateLabels(details.Labels),
						"type":             details.Type,
						"mapTypes":         mapTypes,
						"mapKeys":          mapKeys,
						"mapIKeys":         mapIKeys,
						"global":           details.Global,
						"args":             strings.Join(params[1:], ", "),
						"fparams":          strings.Join(params, ", "),
						"fkeys":            strings.Join(keys, ", "),
					}); err != nil {
						return err
					}

					if err := out.Close(); err != nil {
						return err
					}

					return nil
				}

				if err := save(i, path.Join(root, fmt.Sprintf("%s.go", mname))); err != nil {
					return err
				}

				if err := save(t, path.Join(root, fmt.Sprintf("%s_test.go", mname))); err != nil {
					return err
				}
			}
		}
	}

	out, err := os.OpenFile(path.Join(root, "metrics.go"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	q, err := template.New("metrics").Parse(string(metricsGoTemplate))
	if err != nil {
		return err
	}

	if err := q.Execute(out, nil); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}
