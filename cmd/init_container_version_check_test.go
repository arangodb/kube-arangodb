//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func saveVersionFile(t *testing.T, v int, updates ...func(in *cmdVersionCheckInitContainersInputStruct)) *cmdVersionCheckInitContainersInputStruct {
	var q cmdVersionCheckInitContainersData

	q.Version = v

	d, err := json.Marshal(q)
	require.NoError(t, err)

	var n cmdVersionCheckInitContainersInputStruct

	n.versionPath = fmt.Sprintf("%s/VERSION", t.TempDir())

	require.NoError(t, os.WriteFile(n.versionPath, d, 0644))

	for _, u := range updates {
		u(&n)
	}

	return &n
}

func Test_extractVersionFromData(t *testing.T) {
	check := func(valid bool, name string, version int, updates ...func(in *cmdVersionCheckInitContainersInputStruct)) {
		t.Run(name, func(t *testing.T) {
			err := saveVersionFile(t, version, updates...).Run(nil, nil)
			if valid {
				require.NoError(t, err)
			} else {
				ensureExitCode(t, err, cmdVersionCheckInitContainersInvalidVersionExitCode)
			}
		})
	}

	check(true, "3.9.10_optional", 30910)

	check(true, "3.9.10_required_major", 30910, func(in *cmdVersionCheckInitContainersInputStruct) {
		in.major = 3
	})

	check(true, "3.9.10_required_minor", 30910, func(in *cmdVersionCheckInitContainersInputStruct) {
		in.major = 3
		in.minor = 9
	})

	check(false, "3.9.10_required_major_mismatch", 30910, func(in *cmdVersionCheckInitContainersInputStruct) {
		in.major = 4
	})

	check(false, "3.9.10_required_minor_mismatch", 30910, func(in *cmdVersionCheckInitContainersInputStruct) {
		in.major = 3
		in.minor = 5
	})

	check(true, "3.9.10_required_minor_only_mismatch", 30910, func(in *cmdVersionCheckInitContainersInputStruct) {
		in.minor = 5
	})
}
