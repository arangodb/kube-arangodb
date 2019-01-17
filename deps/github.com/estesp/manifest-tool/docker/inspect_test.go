package docker

import "testing"

func TestSplitHostname(t *testing.T) {
	var crcthostnames = []struct {
		a, b, c string
	}{
		{"localhost:5000/hello-world", "localhost:5000", "hello-world"},
		{"myregistrydomain:5000/java", "myregistrydomain:5000", "java"},
		{"docker.io/busybox", "docker.io", "library/busybox"},
	}
	var wrnghostnames = []struct {
		d, e, f string
	}{
		{"localhost:5000,hello-world", "localhost:5000", "hello-world"},
		{"myregistrydomain:5000&java", "myregistrydomain:5000", "java"},
		{"docker.io%busybox", "docker.io", "busybox"},
	}

	for _, i := range crcthostnames {
		res1, res2 := splitHostname(i.a)
		if res1 != i.b || res2 != i.c {
			t.Errorf("%s should produce equals of: %q %q or %q %q", i.a, res1, i.b, res2, i.c)
		}
	}

	for _, j := range wrnghostnames {
		res1, res2 := splitHostname(j.d)
		if res1 == j.e && res2 == j.f {
			t.Errorf("%s should not produce equals of: %q %q and %q %q", j.d, res1, j.e, res2, j.f)
		}
	}
}

func TestValidateName(t *testing.T) {
	var crctnames = []struct {
		a string
	}{
		{"localhost:5000/hello-world"},
		{"myregistrydomain:5000/java"},
		{"docker.io/busybox"},
	}
	var wrngnames = []struct {
		b string
	}{
		{"localhost:5000,hello-world"},
		{"myregistrydomain:5000&java"},
		{"docker.io@busybox"},
	}

	for _, i := range crctnames {
		res := validateName(i.a)
		if res != nil {
			t.Errorf("%s is an invalid name", i.a)
		}
	}
	for _, j := range wrngnames {
		res := validateName(j.b)
		if res == nil {
			t.Errorf("%s is an invalid name", j.b)
		}
	}
}
