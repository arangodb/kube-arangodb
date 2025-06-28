//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package suite

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	goStrings "strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func RewriteChartName(t *testing.T, old, version, new string) []byte {
	gzIn, err := gzip.NewReader(bytes.NewReader(GetChart(t, old, version)))
	require.NoError(t, err)

	tarIn := tar.NewReader(gzIn)

	out := &bytes.Buffer{}

	gzOut := gzip.NewWriter(out)

	tarOut := tar.NewWriter(gzOut)

	for {
		header, err := tarIn.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		header.Name = fmt.Sprintf("%s%s", new, goStrings.TrimPrefix(header.Name, old))

		switch header.Typeflag {
		case tar.TypeReg:
			if header.Name == fmt.Sprintf("%s/Chart.yaml", new) {
				data, err := io.ReadAll(tarIn)
				require.NoError(t, err)

				var z map[string]interface{}

				require.NoError(t, yaml.Unmarshal(data, &z))

				z["name"] = new

				data, err = yaml.Marshal(z)
				require.NoError(t, err)

				header.Size = int64(len(data))

				require.NoError(t, tarOut.WriteHeader(header))
				_, err = tarOut.Write(data)
				require.NoError(t, err)
			} else {
				require.NoError(t, tarOut.WriteHeader(header))
				_, err = io.Copy(tarOut, tarIn)
				require.NoError(t, err)
			}
		default:
			require.NoError(t, tarOut.WriteHeader(header))
		}
	}

	require.NoError(t, tarOut.Close())
	require.NoError(t, gzOut.Close())
	require.NoError(t, gzIn.Close())

	return out.Bytes()
}
