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

package platform

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
)

type out int

func render(cmd *cobra.Command, f string, args ...interface{}) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), f, args...)
	return err
}

func renderOutput[T any](cmd *cobra.Command, in pretty.Table[T]) error {
	v, err := flagOutput.Get(cmd)
	if err != nil {
		return err
	}
	switch v {
	case "table":
		v, err := in.Redner()
		if err != nil {
			return err
		}
		return render(cmd, v)
	case "json":
		d, err := json.Marshal(in)
		if err != nil {
			return err
		}

		return render(cmd, string(d))
	case "yaml":
		d, err := json.Marshal(in)
		if err != nil {
			return err
		}

		d, err = yaml.JSONToYAML(d)
		if err != nil {
			return err
		}

		return render(cmd, string(d))
	default:
		return errors.Errorf("Unable to render in %s", v)
	}
}
