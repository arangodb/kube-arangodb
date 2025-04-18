//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package pretty

import (
	"bytes"
	"fmt"
	"os"
	goStrings "strings"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func ReplaceSectionsInFile(path string, sections map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	res, err := ReplaceSections(string(data), sections)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(res), 0644)
}

func ReplaceSections(in string, sections map[string]string) (string, error) {
	for k, v := range sections {
		if n, err := ReplaceSection(in, v, k); err != nil {
			return "", err
		} else {
			in = n
		}
	}

	return in, nil
}

func ReplaceSection(in, replace, section string) (string, error) {
	start, end := fmt.Sprintf("\n[START_INJECT]: # (%s)\n", section), fmt.Sprintf("[END_INJECT]: # (%s)\n", section)

	b := bytes.NewBuffer(nil)

	for len(in) > 0 {
		startID := goStrings.Index(in, start)
		if startID == -1 {
			b.WriteString(in)
			in = ""
			continue
		}

		b.WriteString(in[0:startID])

		in = moveString(in, startID+len(start))

		b.WriteString(start)

		b.WriteString(replace)

		endID := goStrings.Index(in, end)
		if endID == -1 {
			return "", errors.Errorf("END_INJECT sections is missing for section %s. Note that newline is required at the end and before tag", section)
		}

		in = moveString(in, endID+len(end))

		b.WriteString(end)

	}

	return b.String(), nil
}

func moveString(in string, offset int) string {
	if offset >= len(in) {
		return ""
	}

	return in[offset:]
}
