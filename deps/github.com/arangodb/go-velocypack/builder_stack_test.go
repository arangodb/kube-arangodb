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

import "testing"

func TestBuilderStack1(t *testing.T) {
	var b builderStack
	if empty := b.IsEmpty(); !empty {
		t.Errorf("Expected empty, got %v", empty)
	}
	b.Push(1)
	if tos, _ := b.Tos(); tos != 1 {
		t.Errorf("Expected 1, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Push(17)
	if tos, _ := b.Tos(); tos != 17 {
		t.Errorf("Expected 17, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Pop()
	if tos, _ := b.Tos(); tos != 1 {
		t.Errorf("Expected 1, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Push(77)
	if tos, _ := b.Tos(); tos != 77 {
		t.Errorf("Expected 77, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Push(88)
	if tos, _ := b.Tos(); tos != 88 {
		t.Errorf("Expected 88, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Pop()
	if tos, _ := b.Tos(); tos != 77 {
		t.Errorf("Expected 77, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Pop()
	if tos, _ := b.Tos(); tos != 1 {
		t.Errorf("Expected 1, got %d", tos)
	}
	if empty := b.IsEmpty(); empty {
		t.Errorf("Expected not empty, got %v", empty)
	}
	b.Pop() // Now empty
	if tos, _ := b.Tos(); tos != 0 {
		t.Errorf("Expected 0, got %d", tos)
	}
	if empty := b.IsEmpty(); !empty {
		t.Errorf("Expected empty, got %v", empty)
	}
	b.Pop() // Already empty
	if tos, _ := b.Tos(); tos != 0 {
		t.Errorf("Expected 0, got %d", tos)
	}
	if empty := b.IsEmpty(); !empty {
		t.Errorf("Expected empty, got %v", empty)
	}
}
