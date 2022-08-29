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

package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func Test_Worker_Empty(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)

	stopCh := make(chan struct{})

	item := randomItem()

	// Act
	require.NoError(t, o.Start(0, stopCh))

	err := o.ProcessItem(item)

	close(stopCh)

	// Assert
	assert.NoError(t, err)
}

func Test_Worker_CatchAll(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)

	stopCh := make(chan struct{})

	item := randomItem()

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	// Act
	require.NoError(t, o.Start(0, stopCh))

	err := o.ProcessItem(item)

	close(stopCh)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, i, 1)

	receivedItem := <-i
	assert.Equal(t, item, receivedItem)

	close(i)
}

func Test_Worker_EnsureFirstProcessStopLoop(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)

	stopCh := make(chan struct{})

	item := randomItem()

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	m2, i2 := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m2))

	// Act
	require.NoError(t, o.Start(0, stopCh))

	err := o.ProcessItem(item)

	close(stopCh)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, i, 1)
	assert.Len(t, i2, 0)

	receivedItem := <-i
	assert.Equal(t, item, receivedItem)

	close(i)
	close(i2)
}

func Test_Worker_EnsureObjectIsProcessedBySecondHandler(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)

	stopCh := make(chan struct{})

	item := randomItem()

	m, i := mockSimpleObject(name, false)
	require.NoError(t, o.RegisterHandler(m))

	m2, i2 := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m2))

	// Act
	require.NoError(t, o.Start(0, stopCh))

	err := o.ProcessItem(item)

	close(stopCh)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, i, 0)
	assert.Len(t, i2, 1)

	receivedItem := <-i2
	assert.Equal(t, item, receivedItem)

	close(i)
	close(i2)
}
