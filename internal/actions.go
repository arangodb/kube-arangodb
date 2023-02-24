//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"os"
	"path"
	"sort"
	"text/template"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/internal/md"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

//go:embed actions.yaml
var actions []byte

//go:embed actions.go.tmpl
var actionsGoTemplate []byte

//go:embed actions.register.go.tmpl
var actionsRegisterGoTemplate []byte

//go:embed actions.register.test.go.tmpl
var actionsRegisterTestGoTemplate []byte

type ActionsInput struct {
	DefaultTimeout meta.Duration `json:"default_timeout"`

	Actions map[string]Action `json:"actions"`
}

func (i ActionsInput) Keys() []string {
	z := make([]string, 0, len(i.Actions))

	for k := range i.Actions {
		z = append(z, k)
	}

	sort.Strings(z)

	return z
}

func (i ActionsInput) Optionals() map[string]bool {
	r := map[string]bool{}

	for k, v := range i.Actions {
		r[k] = v.Optional
	}

	return r
}

type Scopes struct {
	Normal, High, Resource bool
}

func (s Scopes) String() string {
	q := make([]string, 0, 3)
	if s.High {
		q = append(q, "High")
	}
	if s.Normal {
		q = append(q, "Normal")
	}
	if s.Resource {
		q = append(q, "Resource")
	}

	if len(q) > 2 {
		q = []string{
			strings.Join(q[0:len(q)-1], ", "),
			q[len(q)-1],
		}
	}

	return strings.Join(q, " and ")
}

func (i ActionsInput) Scopes() map[string]Scopes {
	r := map[string]Scopes{}
	for k, a := range i.Actions {
		r[k] = Scopes{
			Normal:   a.InScope("normal"),
			High:     a.InScope("high"),
			Resource: a.InScope("resource"),
		}
	}
	return r
}

func (i ActionsInput) StartFailureGracePeriods() map[string]string {
	r := map[string]string{}
	for k, a := range i.Actions {
		if a.StartupFailureGracePeriod == nil {
			r[k] = ""
		} else {
			r[k] = fmt.Sprintf("%d * time.Second", a.StartupFailureGracePeriod.Duration/time.Second)
		}
	}
	return r
}

func (i ActionsInput) Internal() map[string]string {
	r := map[string]string{}

	for a, spec := range i.Actions {
		if spec.IsInternal {
			r[a] = "true"
		}
	}

	return r
}

func (i ActionsInput) HighestScopes() map[string]string {
	r := map[string]string{}
	for k, a := range i.Scopes() {
		if a.High {
			r[k] = "High"
		} else if a.Normal {
			r[k] = "Normal"
		} else if a.Resource {
			r[k] = "Resource"
		} else {
			r[k] = "Unknown"
		}
	}
	return r
}

func (i ActionsInput) Descriptions() map[string]string {
	r := map[string]string{}
	for k, a := range i.Actions {
		r[k] = a.Description
	}
	return r
}

func (i ActionsInput) Timeouts() map[string]string {
	r := map[string]string{}
	for k, a := range i.Actions {
		if a.Timeout != nil {
			r[k] = fmt.Sprintf("%d * time.Second // %s", a.Timeout.Duration/time.Second, a.Timeout.Duration.String())
		} else {
			r[k] = "ActionsDefaultTimeout"
		}
	}
	return r
}

type Action struct {
	Timeout                   *meta.Duration `json:"timeout,omitempty"`
	StartupFailureGracePeriod *meta.Duration `json:"startupFailureGracePeriod,omitempty"`

	Scopes []string `json:"scopes,omitempty"`

	Description string `json:"description"`

	Enterprise bool `json:"enterprise"`

	IsInternal bool `json:"isInternal"`

	Optional bool `json:"optional"`
}

func (a Action) InScope(scope string) bool {
	if a.Scopes == nil {
		return strings.Title(scope) == "Normal"
	}

	for _, x := range a.Scopes {
		if strings.Title(scope) == strings.Title(x) {
			return true
		}
	}

	return false
}

func RenderActions(root string) error {
	var in ActionsInput

	if err := yaml.Unmarshal(actions, &in); err != nil {
		return err
	}

	{
		actions := path.Join(root, "pkg", "apis", "deployment", "v1", "actions.generated.go")

		out, err := os.OpenFile(actions, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		i, err := template.New("actions").Parse(string(actionsGoTemplate))
		if err != nil {
			return err
		}

		if err := i.Execute(out, map[string]interface{}{
			"actions":        in.Keys(),
			"scopes":         in.Scopes(),
			"highestScopes":  in.HighestScopes(),
			"internal":       in.Internal(),
			"timeouts":       in.Timeouts(),
			"descriptions":   in.Descriptions(),
			"optionals":      in.Optionals(),
			"defaultTimeout": fmt.Sprintf("%d * time.Second // %s", in.DefaultTimeout.Duration/time.Second, in.DefaultTimeout.Duration.String()),
		}); err != nil {
			return err
		}

		if err := out.Close(); err != nil {
			return err
		}
	}

	{
		actions := path.Join(root, "pkg", "deployment", "reconcile", "action.register.generated.go")

		out, err := os.OpenFile(actions, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		i, err := template.New("actions").Parse(string(actionsRegisterGoTemplate))
		if err != nil {
			return err
		}

		if err := i.Execute(out, map[string]interface{}{
			"actions":                    in.Keys(),
			"startupFailureGracePeriods": in.StartFailureGracePeriods(),
		}); err != nil {
			return err
		}

		if err := out.Close(); err != nil {
			return err
		}
	}

	{
		actions := path.Join(root, "pkg", "deployment", "reconcile", "action.register.generated_test.go")

		out, err := os.OpenFile(actions, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		i, err := template.New("actions").Parse(string(actionsRegisterTestGoTemplate))
		if err != nil {
			return err
		}

		if err := i.Execute(out, map[string]interface{}{
			"actions":                    in.Keys(),
			"startupFailureGracePeriods": in.StartFailureGracePeriods(),
			"internal":                   in.Internal(),
			"optional":                   in.Optionals(),
		}); err != nil {
			return err
		}

		if err := out.Close(); err != nil {
			return err
		}
	}

	{
		actions := path.Join(root, "docs", "generated", "actions.md")

		action := md.NewColumn("Action", md.ColumnCenterAlign)
		timeout := md.NewColumn("Timeout", md.ColumnCenterAlign)
		description := md.NewColumn("Description", md.ColumnCenterAlign)
		internal := md.NewColumn("Internal", md.ColumnCenterAlign)
		optional := md.NewColumn("Optional", md.ColumnCenterAlign)
		edition := md.NewColumn("Edition", md.ColumnCenterAlign)
		t := md.NewTable(
			action,
			internal,
			timeout,
			optional,
			edition,
			description,
		)

		for _, k := range in.Keys() {
			a := in.Actions[k]
			v := in.DefaultTimeout.Duration.String()
			if t := a.Timeout; t != nil {
				v = t.Duration.String()
			}

			vr := "Community & Enterprise"
			if a.Enterprise {
				vr = "Enterprise Only"
			}
			int := "yes"
			if !a.IsInternal {
				int = "no"
			}
			opt := "yes"
			if !a.Optional {
				opt = "no"
			}

			if err := t.AddRow(map[md.Column]string{
				action:      k,
				timeout:     v,
				description: a.Description,
				edition:     vr,
				optional:    opt,
				internal:    int,
			}); err != nil {
				return err
			}
		}

		timeouts := api.ActionTimeouts{}

		for _, k := range in.Keys() {
			a := in.Actions[k]
			if a.Timeout != nil {
				timeouts[api.ActionType(k)] = api.NewTimeout(a.Timeout.Duration)
			} else {
				timeouts[api.ActionType(k)] = api.NewTimeout(in.DefaultTimeout.Duration)
			}
		}

		d, err := yaml.Marshal(map[string]interface{}{
			"spec": map[string]interface{}{
				"timeouts": map[string]interface{}{
					"actions": timeouts,
				},
			},
		})
		if err != nil {
			return err
		}

		if err := md.ReplaceSectionsInFile(actions, map[string]string{
			"actionsTable":   md.WrapWithNewLines(t.Render()),
			"actionsModYaml": md.WrapWithNewLines(md.WrapWithYAMLSegment(string(d))),
		}); err != nil {
			return err
		}
	}

	return nil
}
