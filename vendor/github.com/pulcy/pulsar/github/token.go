package github

import (
	"io/ioutil"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// LoadGithubToken tries to load a github access token from:
// - GITHUB_TOKEN environment variable
// - ~/.pulcy/github-token file
func LoadGithubToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token, nil
	}
	path, err := homedir.Expand("~/.pulcy/github-token")
	if err != nil {
		return "", maskAny(err)
	}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return "", maskAny(err)
	}
	token = string(raw)
	token = strings.TrimSpace(token)
	return token, nil
}
