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

package v1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Object_Validate(t *testing.T) {
	var o *Object
	require.Error(t, o.Validate())

	o = &Object{}
	require.Error(t, o.Validate())

	o.Name = "#invalid"
	require.Error(t, o.Validate())

	o.Name = "valid"
	require.NoError(t, o.Validate())

	o.Namespace = util.NewType("")
	require.Error(t, o.Validate())

	o.Namespace = util.NewType("#invalid")
	require.Error(t, o.Validate())

	o.Namespace = util.NewType("valid")
	require.NoError(t, o.Validate())
}
