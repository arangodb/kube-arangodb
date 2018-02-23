//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package velocypack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderBufferEmpty(t *testing.T) {
	var b builderBuffer
	assert.Equal(t, ValueLength(0), b.Len())
	assert.True(t, b.IsEmpty())
}

func TestBuilderBufferWriteByte(t *testing.T) {
	var b builderBuffer
	b.WriteByte(5)
	assert.Equal(t, ValueLength(1), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, []byte{5}, b.Bytes())
}

func TestBuilderBufferWriteBytes(t *testing.T) {
	var b builderBuffer
	b.WriteBytes(3, 7)
	assert.Equal(t, ValueLength(7), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, []byte{3, 3, 3, 3, 3, 3, 3}, b.Bytes())
}

func TestBuilderBufferWrite(t *testing.T) {
	var b builderBuffer
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.Write(data)
	assert.Equal(t, ValueLength(len(data)), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, data, b.Bytes())
}

func TestBuilderBufferReserveSpace(t *testing.T) {
	var b builderBuffer
	b.ReserveSpace(32)
	assert.Equal(t, ValueLength(0), b.Len())
	assert.True(t, b.IsEmpty())
	data := []byte{1, 2, 3, 4}
	b.Write(data)
	assert.Equal(t, ValueLength(len(data)), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, data, b.Bytes())
}

func TestBuilderBufferShrink(t *testing.T) {
	var b builderBuffer
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.Write(data)
	assert.Equal(t, ValueLength(len(data)), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, data, b.Bytes())
	b.Shrink(3)
	assert.Equal(t, ValueLength(len(data)-3), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, data[:len(data)-3], b.Bytes())
}

func TestBuilderBufferGrow(t *testing.T) {
	var b builderBuffer
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.Write(data)
	assert.Equal(t, ValueLength(len(data)), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, data, b.Bytes())

	data2 := []byte{5, 6, 7, 8}
	dst := b.Grow(uint(len(data2)))
	copy(dst, data2)
	assert.Equal(t, ValueLength(len(data)+len(data2)), b.Len())
	assert.False(t, b.IsEmpty())
	assert.Equal(t, append(data, data2...), b.Bytes())
}
