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

package get

import (
	"os"
	"path/filepath"

	"github.com/juju/errgo"
	"github.com/mgutz/ansi"
	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/cache"
	"github.com/pulcy/pulsar/git"
	"github.com/pulcy/pulsar/golang"
	"github.com/pulcy/pulsar/util"
)

const (
	defaultGetBranch = "master"
)

var (
	maskAny   = errgo.MaskFunc(errgo.Any)
	allGood   = ansi.ColorFunc("")
	updating  = ansi.ColorFunc("cyan")
	attention = ansi.ColorFunc("yellow")
)

type Flags struct {
	Folder    string
	RepoUrl   string
	Version   string
	AllowLink bool
}

// Get ensures that flags.Folder contains an up to date copy of flags.RepoUrl checked out to flags.Version.
func Get(log *log.Logger, flags *Flags) error {
	// Get cache dir
	cachedir, cacheIsValid, err := cache.Dir(flags.RepoUrl, 0)
	if err != nil {
		return maskAny(err)
	}

	// Expand folder
	flags.Folder, err = filepath.Abs(flags.Folder)
	if err != nil {
		return maskAny(err)
	}

	// Get current folder
	wd, _ := os.Getwd()

	linked := false
	if flags.AllowLink {
		if info, err := util.ParseVCSURL(flags.RepoUrl); err == nil {
			siblingPath := filepath.Join(filepath.Dir(wd), info.Name)
			if _, err := os.Stat(siblingPath); err == nil {
				//log.Infof("Sibling folder %s exists", siblingPath)
				util.ExecuteInDir(siblingPath, func() error {
					remote, err := git.GetRemoteOriginUrl(nil)
					if err != nil {
						return maskAny(err)
					}
					if remote == flags.RepoUrl {
						if relPath, err := filepath.Rel(filepath.Dir(flags.Folder), siblingPath); err == nil {
							if err := os.Symlink(relPath, flags.Folder); err == nil {
								log.Infof("Linked -> %s", siblingPath)
								linked = true

								if vendorDir, err := golang.GetVendorDir(siblingPath); err != nil {
									return maskAny(err)
								} else {
									// Flatten sibling in copy-only mode
									if err := golang.Flatten(log, &golang.FlattenFlags{
										VendorDir: vendorDir,
										NoRemove:  true,
									}); err != nil {
										return maskAny(err)
									}
								}

							}
						}
					}
					return nil
				})
			}
		}
	}
	if linked {
		return nil
	}

	// Fill cache if needed
	cloned := false
	if !cacheIsValid {
		// Clone repo into cachedir
		if err := git.Clone(log, flags.RepoUrl, cachedir); err != nil {
			return maskAny(err)
		}
		cloned = true
	}

	// Make sure a clone exists
	_, err = os.Stat(flags.Folder)
	if os.IsNotExist(err) {
		// Sync into target folder
		if err := os.MkdirAll(flags.Folder, 0777); err != nil {
			return maskAny(err)
		}
		if err := util.ExecPrintError(nil, "rsync", "-a", appendDirSep(cachedir), appendDirSep(flags.Folder)); err != nil {
			return maskAny(err)
		}
	}
	// Change dir to folder
	if err := os.Chdir(flags.Folder); err != nil {
		return maskAny(err)
	}
	// Specific version needed?
	if flags.Version == "" {
		// Get latest version
		if !cloned {
			localCommit, err := git.GetLatestLocalCommit(nil, flags.Folder, defaultGetBranch, false)
			if err != nil {
				return maskAny(err)
			}
			remoteCommit, err := git.GetLatestRemoteCommit(nil, flags.RepoUrl, defaultGetBranch)
			if err != nil {
				return maskAny(err)
			}
			if localCommit != remoteCommit {
				if err := git.Pull(log, "origin"); err != nil {
					return maskAny(err)
				}
			} else {
				log.Info(allGood("%s is up to date\n"), makeRel(wd, flags.Folder))
			}
		}
	} else {
		// Get latest (local) version
		localVersion, err := git.GetLatestTag(nil, flags.Folder)
		if err != nil {
			return maskAny(err)
		}
		if localVersion != flags.Version {
			// Checkout requested version
			if cloned {
				log.Info(updating("Checking out version %s in %s.\n"), flags.Version, makeRel(wd, flags.Folder))
			} else {
				log.Info(updating("Found version %s, wanted %s. Updating %s now.\n"), localVersion, flags.Version, makeRel(wd, flags.Folder))
			}
			// Fetch latest changes
			if err := git.Fetch(log, "origin"); err != nil {
				return maskAny(err)
			}
			if err := git.FetchTags(log, "origin"); err != nil {
				return maskAny(err)
			}
			// Checkout intended version
			if err := git.Checkout(log, flags.Version); err != nil {
				return maskAny(err)
			}
		} else {
			log.Info(allGood("Found correct version. No changes needed in %s\n"), makeRel(wd, flags.Folder))
		}
		// Get latest remote version
		remoteVersion, err := git.GetLatestRemoteTag(nil, flags.RepoUrl)
		if err != nil {
			return maskAny(err)
		}
		if remoteVersion != flags.Version {
			log.Warning(attention("Update available for %s: '%s' => '%s'\n"), makeRel(wd, flags.Folder), flags.Version, remoteVersion)
		}
	}
	return nil
}

// makeRel tries to make the given path relative to the current directory.
// Returns a full path in case of errors.
func makeRel(wd, path string) string {
	rel, err := filepath.Rel(wd, path)
	if err != nil {
		return path
	}
	return rel
}

func appendDirSep(dir string) string {
	if dir == "" {
		return dir
	}
	if dir[len(dir)-1:] != "/" {
		return dir + "/"
	}
	return dir
}
