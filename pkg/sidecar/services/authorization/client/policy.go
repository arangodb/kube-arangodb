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

func newPolicy(in *sidecarSvcAuthzTypes.Policy) (policy, error) {
	var p policy

	if statements, err := util.FormatListErr(in.GetStatements(), func(a *sidecarSvcAuthzTypes.PolicyStatement) (statement, error) {
		return newStatement(a)
	}); err != nil {
		return policy{}, err
	} else {
		p.statements = statements
	}

	return p, nil
}

type policy struct {
	statements []statement
}

func (p *policy) Evaluate(action, resource string, context map[string][]string) (bool, error) {
	if p == nil {
		return false, nil
	}

	var allowed bool

	for _, stmt := range p.statements {
		if stmt.Evaluate(action, resource, context) {
			if stmt.effect == sidecarSvcAuthzTypes.Effect_Deny {
				return false, sidecarSvcAuthzTypes.PermissionDenied{}
			}
			allowed = true
		}
	}

	return allowed, nil
}

func newStatement(in *sidecarSvcAuthzTypes.PolicyStatement) (statement, error) {
	var s statement
	s.effect = in.GetEffect()

	if actions, err := util.FormatListErr(in.GetActions(), func(a string) (match, error) {
		return newAction(a)
	}); err != nil {
		return statement{}, err
	} else {
		s.actions = actions
	}

	if resources, err := util.FormatListErr(in.GetResources(), func(a string) (match, error) {
		return newMatch(a)
	}); err != nil {
		return statement{}, err
	} else {
		s.resources = resources
	}

	return s, nil
}

type statement struct {
	effect    sidecarSvcAuthzTypes.Effect
	resources matches
	actions   matches
}

func (p *statement) Evaluate(action, resource string, context map[string][]string) bool {
	if p == nil {
		return false
	}

	// Skip context

	return p.actions.match(action) && p.resources.match(resource)
}

type matches []match

func (m matches) match(resource string) bool {
	for _, match := range m {
		if match.match(resource) {
			return true
		}
	}
	return false
}

func newAction(in string) (match, error) {
	p, err := newMultiMatch(in)
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

func newMultiMatch(v string) (match, error) {
	r := goStrings.Split(v, ":")

	q, err := util.FormatListErr(r, func(a string) (match, error) {
		return newMatch(a)
	})
	if err != nil {
		return nil, err
	}

	if len(q) == 1 && isAllMatch(q[0]) {
		return q[0], nil
	}

	return manyMatch(q), nil
}

func newMatch(v string) (match, error) {
	if v == "*" {
		return allMatch{}, nil
	}

	if suffix, prefix := goStrings.HasSuffix(v, "*"), goStrings.HasPrefix(v, "*"); suffix && prefix {
		return nil, errors.Errorf("invalid policy statement: %s", v)
	} else if suffix {
		return suffixMatch(goStrings.TrimSuffix(v, "*")), nil
	} else if prefix {
		return prefixMatch(goStrings.TrimPrefix(v, "*")), nil
	}

	return exactMatch(v), nil
}

type match interface {
	match(resource string) bool
}

type manyMatch []match

func (m manyMatch) match(resource string) bool {
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

		if !m[id].match(res[id]) {
			return false
		}
	}

	return true
}

type exactMatch string

func (e exactMatch) match(resource string) bool {
	return string(e) == resource
}

func isAllMatch(in match) bool {
	_, ok := in.(allMatch)
	return ok
}

type allMatch struct{}

func (a allMatch) match(resource string) bool {
	return true
}

type prefixMatch string

func (p prefixMatch) match(resource string) bool {
	return goStrings.HasSuffix(resource, string(p))
}

type suffixMatch string

func (s suffixMatch) match(resource string) bool {
	return goStrings.HasPrefix(resource, string(s))
}
