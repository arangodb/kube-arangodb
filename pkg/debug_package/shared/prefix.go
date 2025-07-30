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

package shared

import (
	"fmt"
	"reflect"

	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func WithSlash(files chan<- File) (chan<- File, func()) {
	return WithPrefix(files, "/")
}

func WithGVRPrefix(files chan<- File, t reflect.Type) (chan<- File, func()) {
	gvr, ok := inspectorConstants.ExtractGVR(t)
	if !ok {
		panic(fmt.Sprintf("Unable to get GVR for %s", t.String()))
	}

	if gvr.Group == "" {
		gvr.Group = "core"
	}

	return WithPrefix(files, "%s/%s/%s", gvr.Group, gvr.Version, gvr.Resource)
}

func WithPrefix(files chan<- File, f string, args ...any) (chan<- File, func()) {
	done := make(chan any)

	r := make(chan File)

	go func() {
		defer close(done)

		for el := range r {
			files <- prefixedFile{
				prefix: fmt.Sprintf(f, args...),
				up:     el,
			}
		}
	}()

	return r, func() {
		close(r)

		<-done
	}
}

type prefixedFile struct {
	prefix string
	up     File
}

func (p prefixedFile) Path() string {
	return fmt.Sprintf("%s%s", p.prefix, p.up.Path())
}

func (p prefixedFile) Write() ([]byte, error) {
	return p.up.Write()
}
