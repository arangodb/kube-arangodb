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

package internal

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"
)

//go:embed timezones.go.tmpl
var timezonesGoTemplate []byte

type Timezone struct {
	Name   string
	Offset int64
	Zone   string
}

func RenderTimezones(root string) error {
	rootPath := path.Join(root, "pkg", "generated", "timezones")

	if _, err := os.Stat(rootPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(rootPath, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	out, err := os.OpenFile(path.Join(rootPath, "timezones.go"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	i, err := template.New("timezones").Parse(string(timezonesGoTemplate))
	if err != nil {
		return err
	}

	if err := i.Execute(out, map[string]interface{}{
		"timezones": ListTimezones(),
	}); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

func ListTimezones() []Timezone {
	var zoneDirs = []string{
		// Update path according to your OS
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}

	zones := map[string]time.Time{}

	now := time.Now()

	var tzs []Timezone

	for _, zoneDir := range zoneDirs {
		files, err := ioutil.ReadDir(zoneDir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if !file.IsDir() {
				continue
			}

			fn := file.Name()
			if fn[0] != strings.ToUpper(fn)[0] {
				continue
			}

			if fn[1:] != strings.ToLower(fn)[1:] {
				continue
			}

			subFiles, err := ioutil.ReadDir(path.Join(zoneDir, fn))
			if err != nil {
				continue
			}

			for _, subFile := range subFiles {
				zn := fmt.Sprintf("%s/%s", fn, subFile.Name())

				loc, err := time.LoadLocation(zn)
				if err != nil {
					continue
				}

				zones[zn] = now.In(loc)
			}
		}
	}

	for tz, t := range zones {
		q, o := t.Zone()
		tzs = append(tzs, Timezone{
			Name:   tz,
			Offset: (int64(o) * int64(time.Second)) / int64(time.Minute),
			Zone:   q,
		})
	}

	sort.Slice(tzs, func(i, j int) bool {
		return tzs[i].Name < tzs[j].Name
	})

	return tzs
}
