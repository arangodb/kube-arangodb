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

package authentication

import (
	"context"
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

type SigningKeys interface {
	Enabled() bool

	Get(ctx context.Context) ([]string, string, error)
}

type secretSigningKeys struct {
	secret cache.Object[utilToken.Secret]
}

func (s secretSigningKeys) Enabled() bool {
	return true
}

func (s secretSigningKeys) Get(ctx context.Context) ([]string, string, error) {
	secret, err := s.secret.Get(ctx)
	if err != nil {
		return nil, "", err
	}

	pk := secret.PublicKey()

	sort.Strings(pk)

	return util.UniqueList(pk), secret.Hash(), nil
}

type emptySigningKeys struct{}

func (e emptySigningKeys) Enabled() bool {
	return false
}

func (e emptySigningKeys) Get(ctx context.Context) ([]string, string, error) {
	return nil, "", errors.Errorf("not implemented")
}
