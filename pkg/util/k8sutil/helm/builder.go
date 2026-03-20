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

package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewChartBuilder(out io.Writer, d ChartDefinition) ChartBuilder {
	gz := gzip.NewWriter(out)

	tw := tar.NewWriter(gz)
	r := &chartBuilder{
		spec: d,
		gzip: gz,
		tar:  tw,
	}

	return r.YAMLFile("Chart.yaml", d)
}

type ChartBuilder interface {
	File(path string, content []byte) ChartBuilder

	YAMLFile(path string, objects ...any) ChartBuilder

	Done() error
}

type ChartDefinition struct {
	ApiVersion  string `json:"apiVersion"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type chartBuilder struct {
	spec ChartDefinition

	lock sync.Mutex

	gzip *gzip.Writer

	tar *tar.Writer
}

func (c *chartBuilder) YAMLFile(path string, objects ...any) ChartBuilder {
	data, err := util.FormatListErr(objects, func(a any) ([]byte, error) {
		return yaml.Marshal(a)
	})
	if err != nil {
		return errorChartBuilder{err: err}
	}

	return c.File(path, bytes.Join(data, []byte("\n---\n\n")))
}

func (c *chartBuilder) File(path string, content []byte) ChartBuilder {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.tar.WriteHeader(&tar.Header{
		Name: fmt.Sprintf("%s/%s", c.spec.Name, path),
		Mode: 0644,
		Uid:  1000,
		Gid:  1000,
		Size: int64(len(content)),
	}); err != nil {
		return errorChartBuilder{err: err}
	}

	if _, err := c.tar.Write(content); err != nil {
		return errorChartBuilder{err: err}
	}

	return c
}

func (c *chartBuilder) Done() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.tar.Close(); err != nil {
		return err
	}

	if err := c.gzip.Close(); err != nil {
		return err
	}

	return nil
}

type errorChartBuilder struct {
	err error
}

func (e errorChartBuilder) YAMLFile(path string, objects ...any) ChartBuilder {
	return e
}

func (e errorChartBuilder) File(path string, content []byte) ChartBuilder {
	return e
}

func (e errorChartBuilder) Done() error {
	return e.err
}
