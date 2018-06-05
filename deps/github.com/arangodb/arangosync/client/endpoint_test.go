//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// The Programs (which include both the software and documentation) contain
// proprietary information of ArangoDB GmbH; they are provided under a license
// agreement containing restrictions on use and disclosure and are also
// protected by copyright, patent and other intellectual and industrial
// property laws. Reverse engineering, disassembly or decompilation of the
// Programs, except to the extent required to obtain interoperability with
// other independently created software or as specified by law, is prohibited.
//
// It shall be the licensee's responsibility to take all appropriate fail-safe,
// backup, redundancy, and other measures to ensure the safe use of
// applications if the Programs are used for purposes such as nuclear,
// aviation, mass transit, medical, or other inherently dangerous applications,
// and ArangoDB GmbH disclaims liability for any damages caused by such use of
// the Programs.
//
// This software is the confidential and proprietary information of ArangoDB
// GmbH. You shall not disclose such confidential and proprietary information
// and shall use it only in accordance with the terms of the license agreement
// you entered into with ArangoDB GmbH.
//
// Author Ewout Prangsma
//

package client

import "testing"

func TestEndpointContains(t *testing.T) {
	ep := Endpoint{"http://a", "http://b", "http://c"}
	for _, x := range []string{"http://a", "http://b", "http://c", "http://a/"} {
		if !ep.Contains(x) {
			t.Errorf("Expected endpoint to contain '%s' but it did not", x)
		}
	}
	for _, x := range []string{"", "http://ab", "-", "http://abc"} {
		if ep.Contains(x) {
			t.Errorf("Expected endpoint to not contain '%s' but it did", x)
		}
	}
}

func TestEndpointIsEmpty(t *testing.T) {
	ep := Endpoint{"http://a", "http://b", "http://c"}
	if ep.IsEmpty() {
		t.Error("Expected endpoint to be not empty, but it is")
	}
	ep = nil
	if !ep.IsEmpty() {
		t.Error("Expected endpoint to be empty, but it is not")
	}
	ep = Endpoint{}
	if !ep.IsEmpty() {
		t.Error("Expected endpoint to be empty, but it is not")
	}
}

func TestEndpointEquals(t *testing.T) {
	expectEqual := []Endpoint{
		Endpoint{}, Endpoint{},
		Endpoint{}, nil,
		Endpoint{"http://a"}, Endpoint{"http://a"},
		Endpoint{"http://a", "http://b"}, Endpoint{"http://b", "http://a"},
		Endpoint{"http://foo:8529"}, Endpoint{"http://foo:8529/"},
	}
	for i := 0; i < len(expectEqual); i += 2 {
		epa := expectEqual[i]
		epb := expectEqual[i+1]
		if !epa.Equals(epb) {
			t.Errorf("Expected endpoint %v to be equal to %v, but it is not", epa, epb)
		}
		if !epb.Equals(epa) {
			t.Errorf("Expected endpoint %v to be equal to %v, but it is not", epb, epa)
		}
	}

	expectNotEqual := []Endpoint{
		Endpoint{"http://a"}, Endpoint{},
		Endpoint{"http://z"}, nil,
		Endpoint{"http://aa"}, Endpoint{"http://a"},
		Endpoint{"http://a:100"}, Endpoint{"http://a:200"},
		Endpoint{"http://a", "http://b", "http://c"}, Endpoint{"http://b", "http://a"},
	}
	for i := 0; i < len(expectNotEqual); i += 2 {
		epa := expectNotEqual[i]
		epb := expectNotEqual[i+1]
		if epa.Equals(epb) {
			t.Errorf("Expected endpoint %v to be not equal to %v, but it is", epa, epb)
		}
		if epb.Equals(epa) {
			t.Errorf("Expected endpoint %v to be not equal to %v, but it is", epb, epa)
		}
	}
}

func TestEndpointClone(t *testing.T) {
	tests := []Endpoint{
		Endpoint{},
		Endpoint{"http://a"},
		Endpoint{"http://a", "http://b"},
	}
	for _, orig := range tests {
		c := orig.Clone()
		if !orig.Equals(c) {
			t.Errorf("Expected endpoint %v to be equal to clone %v, but it is not", orig, c)
		}
		if len(c) > 0 {
			c[0] = "http://modified"
			if orig.Equals(c) {
				t.Errorf("Expected endpoint %v to be no longer equal to clone %v, but it is", orig, c)
			}
		}
	}
}

func TestEndpointIntersection(t *testing.T) {
	expectIntersection := []Endpoint{
		Endpoint{"http://a"}, Endpoint{"http://a"},
		Endpoint{"http://a"}, Endpoint{"http://a", "http://b"},
		Endpoint{"http://a"}, Endpoint{"http://b", "http://a"},
		Endpoint{"http://a", "http://b"}, Endpoint{"http://b", "http://foo27"},
		Endpoint{"http://foo:8529"}, Endpoint{"http://foo:8529/"},
	}
	for i := 0; i < len(expectIntersection); i += 2 {
		epa := expectIntersection[i]
		epb := expectIntersection[i+1]
		if len(epa.Intersection(epb)) == 0 {
			t.Errorf("Expected endpoint %v to have an intersection with %v, but it does not", epa, epb)
		}
		if len(epb.Intersection(epa)) == 0 {
			t.Errorf("Expected endpoint %v to have an intersection with %v, but it does not", epb, epa)
		}
	}

	expectNoIntersection := []Endpoint{
		Endpoint{"http://a"}, Endpoint{},
		Endpoint{"http://z"}, nil,
		Endpoint{"http://aa"}, Endpoint{"http://a"},
		Endpoint{"http://a", "http://b", "http://c"}, Endpoint{"http://e", "http://f"},
	}
	for i := 0; i < len(expectNoIntersection); i += 2 {
		epa := expectNoIntersection[i]
		epb := expectNoIntersection[i+1]
		if len(epa.Intersection(epb)) > 0 {
			t.Errorf("Expected endpoint %v to have no intersection with %v, but it does", epa, epb)
		}
		if len(epb.Intersection(epa)) > 0 {
			t.Errorf("Expected endpoint %v to havenoan intersection with %v, but it does", epb, epa)
		}
	}
}

func TestEndpointValidate(t *testing.T) {
	validTests := []Endpoint{
		Endpoint{},
		Endpoint{"http://a"},
		Endpoint{"http://a", "http://b"},
	}
	for _, x := range validTests {
		if err := x.Validate(); err != nil {
			t.Errorf("Expected endpoint %v to be valid, but it is not because %s", x, err)
		}
	}
	invalidTests := []Endpoint{
		Endpoint{":http::foo"},
		Endpoint{"http/a"},
		Endpoint{"http??"},
		Endpoint{"http:/"},
		Endpoint{"http:/foo"},
	}
	for _, x := range invalidTests {
		if err := x.Validate(); err == nil {
			t.Errorf("Expected endpoint %v to be not valid, but it is", x)
		}
	}
}

func TestEndpointURLs(t *testing.T) {
	ep := Endpoint{"http://a", "http://b/rel"}
	expected := []string{"http://a", "http://b"}
	list, err := ep.URLs()
	if err != nil {
		t.Errorf("URLs expected to succeed, but got %s", err)
	} else {
		for i, x := range list {
			found := x.String()
			if found != expected[i] {
				t.Errorf("Unexpected URL at index %d of %v, expected '%s', got '%s'", i, ep, expected[i], found)
			}
		}
	}
}

func TestEndpointMerge(t *testing.T) {
	tests := []Endpoint{
		Endpoint{"http://a"}, Endpoint{}, Endpoint{"http://a"},
		Endpoint{"http://z"}, nil, Endpoint{"http://z"},
		Endpoint{"http://aa"}, Endpoint{"http://a"}, Endpoint{"http://aa", "http://a"},
		Endpoint{"http://a", "http://b", "http://c"}, Endpoint{"http://e", "http://f"}, Endpoint{"http://a", "http://b", "http://c", "http://e", "http://f"},
		Endpoint{"http://a", "http://b", "http://c"}, Endpoint{"http://a", "http://f"}, Endpoint{"http://a", "http://b", "http://c", "http://f"},
	}
	for i := 0; i < len(tests); i += 3 {
		epa := tests[i]
		epb := tests[i+1]
		expected := tests[i+2]
		result := epa.Merge(epb...)
		if !result.Equals(expected) {
			t.Errorf("Expected merge of endpoints %v & %v to be %v, but got %v", epa, epb, expected, result)
		}
	}
}
