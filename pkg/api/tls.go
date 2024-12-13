//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package api

import (
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func prepareTLSConfig(cli typedCore.CoreV1Interface, cfg ServerConfig) util.TLSConfigFetcher {
	if cfg.TLSSecretName != "" {
		return util.NewSecretTLSConfig(cli.Secrets(cfg.Namespace), cfg.TLSSecretName)
	}

	return util.NewSelfSignedTLSConfig(cfg.ServerName, cfg.ServerAltNames...)
}
