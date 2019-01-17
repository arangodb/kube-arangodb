package docker

import "testing"

func TestValidOSArch(t *testing.T) {
	var crctosarch = []struct {
		arch, os string
	}{
		{"darwin", "386"},
		{"linux", "arm"},
		{"windows", "amd64"},
		{"darwin", "386"},
		{"darwin", "amd64"},
		{"darwin", "arm"},
		{"darwin", "arm64"},
		{"dragonfly", "amd64"},
		{"freebsd", "386"},
		{"freebsd", "amd64"},
		{"freebsd", "arm"},
		{"linux", "386"},
		{"linux", "amd64"},
		{"linux", "arm"},
		{"linux", "arm64"},
		{"linux", "ppc64"},
		{"linux", "ppc64le"},
		{"linux", "mips64"},
		{"linux", "mips64le"},
		{"linux", "s390x"},
		{"netbsd", "386"},
		{"netbsd", "amd64"},
		{"netbsd", "arm"},
		{"openbsd", "386"},
		{"openbsd", "amd64"},
		{"openbsd", "arm"},
		{"plan9", "386"},
		{"plan9", "amd64"},
		{"solaris", "amd64"},
		{"windows", "386"},
		{"windows", "amd64"},
	}
	var wrongosarch = []struct {
		arch, os string
	}{
		{"abc", "123"},
		{"xyz", "etc"},
		{"", ""},
	}

	for _, i := range crctosarch {
		res := isValidOSArch(i.arch, i.os)
		if res != true {
			t.Errorf("%s/%s is an invalid os/arch combination", i.arch, i.os)
		}
	}

	for _, j := range wrongosarch {
		res := isValidOSArch(j.arch, j.os)
		if res == true {
			t.Errorf("%s/%s is an invalid os/arch combination", j.arch, j.os)
		}
	}

}
