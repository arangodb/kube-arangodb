// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"github.com/juju/errgo"
	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/util"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

const (
	cmdName   = "git"
	tagMarker = "refs/tags/"
)

// Execute a `git add`
func Add(log *log.Logger, files ...string) error {
	args := []string{"add"}
	args = append(args, files...)
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git commit`
func Commit(log *log.Logger, message string) error {
	if msg, err := util.Exec(log, cmdName, "commit", "-m", message); err != nil {
		fmt.Printf("%s\n", msg)
		return maskAny(err)
	}
	return nil
}

// Execute a `git status`
func Status(log *log.Logger, porcelain bool) (string, error) {
	args := []string{"status"}
	if porcelain {
		args = append(args, "--porcelain")
	}
	if msg, err := util.Exec(log, cmdName, args...); err != nil {
		if log != nil {
			log.Error(msg)
		} else {
			fmt.Printf("%s\n", msg)
		}
		return "", maskAny(err)
	} else {
		return strings.TrimSpace(msg), nil
	}
}

// Execute a `git status a b`
func Diff(log *log.Logger, a, b string) (string, error) {
	args := []string{"diff",
		a,
		b,
	}
	if msg, err := util.Exec(log, cmdName, args...); err != nil {
		if log != nil {
			log.Error(msg)
		} else {
			fmt.Printf("%s\n", msg)
		}
		return "", maskAny(err)
	} else {
		return strings.TrimSpace(msg), nil
	}
}

// Execute a `git push`
func Push(log *log.Logger, remote string, tags bool) error {
	args := []string{
		"push",
	}
	if tags {
		args = append(args, "--tags")
	}
	if remote != "" {
		args = append(args, remote)
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git pull`
func Pull(log *log.Logger, remote string) error {
	args := []string{
		"pull",
	}
	if remote != "" {
		args = append(args, remote)
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git tag <tag>`
func Tag(log *log.Logger, tag string) error {
	args := []string{
		"tag",
		tag,
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git fetch <remote>`
func Fetch(log *log.Logger, remote string) error {
	args := []string{
		"fetch",
		remote,
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git fetch --tags <remote>`
func FetchTags(log *log.Logger, remote string) error {
	args := []string{
		"fetch",
		"--tags",
		remote,
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Execute a `git clone <repo-url> <folder>`
func Clone(log *log.Logger, repoUrl, folder string) error {
	args := []string{
		"clone",
		repoUrl,
		folder,
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Gets the latest tag from the repo in given folder.
func GetLatestTag(log *log.Logger, folder string) (string, error) {
	args := []string{
		"describe",
		"--abbrev=0",
		"--tags",
	}
	cmd := util.PrepareCommand(log, cmdName, args...)
	cmd.SetDir(folder)
	output, err := cmd.Run()
	if util.IsExit(err) {
		return "", nil
	} else if err != nil {
		return "", maskAny(err)
	}
	return strings.TrimSpace(output), nil
}

// Execute a `git checkout <branch>`
func Checkout(log *log.Logger, branch string) error {
	args := []string{
		"checkout",
		branch,
	}
	return maskAny(util.ExecPrintError(log, cmdName, args...))
}

// Gets the tags from the given remote git repo.
func GetRemoteTags(log *log.Logger, repoUrl string) (TagList, error) {
	args := []string{
		"ls-remote",
		"--tags",
		repoUrl,
	}
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return []string{}, maskAny(err)
	}
	tags := TagList{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		index := strings.Index(line, tagMarker)
		if index < 0 {
			continue
		}
		tag := line[index+len(tagMarker):]
		tags = append(tags, tag)
	}
	if err := scanner.Err(); err != nil {
		return tags, maskAny(err)
	}

	// Sort tags from high to low
	sort.Sort(tags)

	return tags, nil
}

// Gets the latest tags from the given remote git repo.
func GetLatestRemoteTag(log *log.Logger, repoUrl string) (string, error) {
	tags, err := GetRemoteTags(log, repoUrl)
	if err != nil {
		return "", maskAny(err)
	}
	if len(tags) > 0 {
		return tags[0], nil
	}
	return "", nil
}

// Gets the latest commit hash from the given local git folder.
func GetLatestLocalCommit(log *log.Logger, folder, branch string, short bool) (string, error) {
	if branch == "" {
		branch = "HEAD"
	}
	args := []string{"rev-parse"}
	if short {
		args = append(args, "--short")
	}
	args = append(args, branch)
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return "", maskAny(err)
	}
	return strings.TrimSpace(output), nil
}

// Gets the latest commit hash from the given remote git repo + optional branch.
func GetLatestRemoteCommit(log *log.Logger, repoUrl, branch string) (string, error) {
	args := []string{
		"ls-remote",
		repoUrl,
	}
	if branch != "" {
		args = append(args, branch)
	}
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return "", maskAny(err)
	}
	parts := strings.Split(output, "\t")
	return parts[0], nil
}

// Gets the name of the current branch
func GetLocalBranchName(log *log.Logger) (string, error) {
	args := []string{
		"rev-parse",
		"--abbrev-ref",
		"HEAD",
	}
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return "", maskAny(err)
	}
	return strings.TrimSpace(output), nil
}

// Gets the a config value
func GetConfig(log *log.Logger, key string) (string, error) {
	args := []string{
		"config",
		"--get",
		key,
	}
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return "", maskAny(err)
	}
	return strings.TrimSpace(output), nil
}

// Gets the config value for "remote.origin.url"
func GetRemoteOriginUrl(log *log.Logger) (string, error) {
	return GetConfig(log, "remote.origin.url")
}

// git show-branch --merge-base
func GetMergeBase(log *log.Logger) (string, error) {
	args := []string{
		"show-branch",
		"--merge-base",
	}
	output, err := util.Exec(log, cmdName, args...)
	if err != nil {
		return "", maskAny(err)
	}
	return strings.TrimSpace(output), nil
}
