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
	"fmt"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
	utilTokenLoader "github.com/arangodb/kube-arangodb/pkg/util/token/loader"
)

func NewFolderAuthentication(path string, mods ...util.ModR[utilToken.Claims]) Authentication {
	return &FolderAuthentication{
		in:   cache.NewObject(utilTokenLoader.SecretCacheDirectory(path, time.Minute)),
		mods: mods,
	}
}

type FolderAuthentication struct {
	in cache.Object[utilToken.Secret]

	mods []util.ModR[utilToken.Claims]
}

func (s FolderAuthentication) ExtendAuthentication(ctx context.Context) (string, bool, error) {
	Folder, err := s.in.Get(ctx)
	if err != nil {
		return "", false, err
	}

	if !Folder.Exists() {
		return "", false, nil
	}

	token, err := utilToken.NewClaims().With(s.mods...).Sign(Folder)
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("bearer %s", token), true, nil
}
