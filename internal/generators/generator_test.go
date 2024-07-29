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

package generators

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

//go:embed generator_pkg_operatorv2_handlers.go.tmpl
var generatorPKGOperatorV2Handlers []byte

func Test_Generate_PKG_OperatorV2_Handlers(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	i, err := template.New("metrics").Parse(string(generatorPKGOperatorV2Handlers))
	require.NoError(t, err)

	for id := 0; id < 10; id++ {
		t.Run(fmt.Sprintf("%d", id), func(t *testing.T) {
			out, err := os.OpenFile(path.Join(root, "pkg/operatorV2", fmt.Sprintf("handler_p%d.generated.go", id)), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			require.NoError(t, err)

			var params []string

			for z := 0; z < id; z++ {
				params = append(params, fmt.Sprintf("P%d", z+1))
			}

			cleanVars := strings.Join(params, ", ")

			cleanRefs := strings.Join(util.FormatList(params, func(a string) string {
				return strings.ToLower(a)
			}), ", ")

			templateVars := strings.Join(params, ", ")
			templateInputVars := strings.Join(params, ", ")
			inputVars := strings.Join(util.FormatList(params, func(a string) string {
				return fmt.Sprintf("%s %s", strings.ToLower(a), a)
			}), ", ")

			if templateVars != "" {
				templateVars = fmt.Sprintf("[%s any]", templateVars)
			}

			if templateInputVars != "" {
				templateInputVars = fmt.Sprintf("[%s]", templateInputVars)
			}

			require.NoError(t, i.Execute(out, map[string]interface{}{
				"id":                id,
				"templateVars":      templateVars,
				"templateInputVars": templateInputVars,
				"inputVars":         inputVars,
				"cleanVars":         cleanVars,
				"cleanRefs":         cleanRefs,
			}))

			require.NoError(t, out.Close())
		})
	}
}
