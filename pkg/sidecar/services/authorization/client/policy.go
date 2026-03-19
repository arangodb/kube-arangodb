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

package client

import (
	"fmt"
	goStrings "strings"

	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewPolicy(in *sidecarSvcAuthzTypes.Policy) (Policy, error) {
	var p Policy

	if statements, err := util.FormatListErr(in.GetStatements(), func(a *sidecarSvcAuthzTypes.PolicyStatement) (Statement, error) {
		return NewStatement(a)
	}); err != nil {
		return Policy{}, err
	} else {
		p.Statements = statements
	}

	return p, nil
}

type Policy struct {
	Statements []Statement
}

func (p *Policy) Evaluate(action, resource string, context map[string][]string) (bool, error) {
	if p == nil {
		return false, nil
	}

	var allowed bool

	for _, stmt := range p.Statements {
		if stmt.Evaluate(action, resource, context) {
			if stmt.Effect == sidecarSvcAuthzTypes.Effect_Deny {
				return false, sidecarSvcAuthzTypes.PermissionDenied{}
			}
			allowed = true
		}
	}

	return allowed, nil
}

func NewStatement(in *sidecarSvcAuthzTypes.PolicyStatement) (Statement, error) {
	var s Statement
	s.Effect = in.GetEffect()

	if actions, err := util.FormatListErr(in.GetActions(), func(a string) (Match, error) {
		return NewAction(a)
	}); err != nil {
		return Statement{}, err
	} else {
		s.actions = actions
	}

	if resources, err := util.FormatListErr(in.GetResources(), func(a string) (Match, error) {
		return NewMultiMatch(a)
	}); err != nil {
		return Statement{}, err
	} else {
		s.Resources = resources
	}

	return s, nil
}

type Statement struct {
	Effect    sidecarSvcAuthzTypes.Effect
	Resources Matches
	actions   Matches
}

func (p *Statement) Evaluate(action, resource string, context map[string][]string) bool {
	if p == nil {
		return false
	}

	// Skip context

	return p.actions.Match(action) && p.Resources.Match(resource)
}

type Matches []Match

func (m Matches) Match(resource string) bool {
	for _, match := range m {
		if match.Match(resource) {
			return true
		}
	}
	return false
}

func NewAction(in string) (Match, error) {
	p, err := NewMultiMatch(in)
	if err != nil {
		return nil, err
	}

	switch v := p.(type) {
	case allMatch:
		return v, nil
	case manyMatch:
		if len(v) == 2 {
			return p, nil
		}
	}
	return nil, fmt.Errorf("invalid action: %s", in)
}

func NewMultiMatch(v string) (Match, error) {
	r := goStrings.Split(v, ":")

	q, err := util.FormatListErr(r, func(a string) (Match, error) {
		return NewMatch(a)
	})
	if err != nil {
		return nil, err
	}

	if len(q) == 1 && isAllMatch(q[0]) {
		return q[0], nil
	}

	return manyMatch(q), nil
}

func NewMatch(v string) (Match, error) {
	if v == "*" {
		return allMatch{}, nil
	}

	if suffix, prefix := goStrings.HasSuffix(v, "*"), goStrings.HasPrefix(v, "*"); suffix && prefix {
		return nil, errors.Errorf("invalid policy statement: %s", v)
	} else if suffix {
		return prefixMatch(goStrings.TrimSuffix(v, "*")), nil
	} else if prefix {
		return suffixMatch(goStrings.TrimPrefix(v, "*")), nil
	}

	return exactMatch(v), nil
}

type Match interface {
	Match(resource string) bool
}

type manyMatch []Match

func (m manyMatch) Match(resource string) bool {
	res := goStrings.Split(resource, ":")

	if len(m) == 0 {
		return false
	}

	if len(res) < len(m) {
		return false
	} else if len(res) != len(m) {
		if !isAllMatch(m[len(m)-1]) {
			// As last item is not wildcard we wont be able to accept it anyway
			return false
		}
	}

	for id := range res {
		if id >= len(m) {
			// We are at the last element, it was a star so lets move on
			return true
		}

		if !m[id].Match(res[id]) {
			return false
		}
	}

	return true
}

type exactMatch string

func (e exactMatch) Match(resource string) bool {
	return string(e) == resource
}

func isAllMatch(in Match) bool {
	_, ok := in.(allMatch)
	return ok
}

type allMatch struct{}

func (a allMatch) Match(resource string) bool {
	return true
}

type suffixMatch string

func (p suffixMatch) Match(resource string) bool {
	return goStrings.HasSuffix(resource, string(p))
}

type prefixMatch string

func (s prefixMatch) Match(resource string) bool {
	return goStrings.HasPrefix(resource, string(s))
}
