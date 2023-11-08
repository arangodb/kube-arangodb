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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GenerateFeaturesIndex(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	outPath := path.Join(root, "docs/features/README.md")

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, out.Close())
	}()

	const basePath = "docs/features"
	write(t, out, "## List of Community Edition features\n")
	section, err := GenerateReadmeFeatures(root, basePath, false)
	require.NoError(t, err)
	write(t, out, section)
	write(t, out, "\n")

	write(t, out, "## List of Enterprise Edition features\n")
	section, err = GenerateReadmeFeatures(root, basePath, true)
	require.NoError(t, err)
	write(t, out, section)
	write(t, out, "\n")
}
