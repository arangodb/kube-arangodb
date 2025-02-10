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

package cli

import "github.com/spf13/cobra"

type RunE func(cmd *cobra.Command, args []string) error

type Runner []RunE

func (r Runner) Run(cmd *cobra.Command, args []string) error {
	for _, e := range r {
		if e != nil {
			if err := e(cmd, args); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r Runner) With(calls ...RunE) Runner {
	ret := make(Runner, len(r)+len(calls))

	copy(ret, r)

	copy(ret[len(r):], calls)

	return ret
}
