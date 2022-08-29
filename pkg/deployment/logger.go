//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var (
	logger = logging.Global().RegisterAndGetLogger("deployment", logging.Info)
)

func (d *Deployment) sectionLogger(section string) logging.Logger {
	return d.log.Str("section", section)
}

func (d *Deployment) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in.Str("namespace", d.namespace).Str("name", d.name)
}
