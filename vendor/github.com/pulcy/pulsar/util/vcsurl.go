package util

import (
	"fmt"
	"regexp"

	vcsurl "github.com/sourcegraph/go-vcsurl"
)

var (
	gitPreprocessRE = regexp.MustCompile("^git@([a-zA-Z0-9-_\\.]+)\\:(.*)$")
)

func ParseVCSURL(url string) (*vcsurl.RepoInfo, error) {
	if parts := gitPreprocessRE.FindStringSubmatch(url); len(parts) == 3 {
		url = fmt.Sprintf("git://%s/%s", parts[1], parts[2])
	}
	return vcsurl.Parse(url)
}
