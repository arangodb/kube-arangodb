# Pulsar: Pulcy Development Environment

[![Build Status](https://travis-ci.org/pulcy/pulsar.svg?branch=master)](https://travis-ci.org/pulcy/pulsar)

## Requirements

* Docker
* Git
* Go
* Node.js (optional)
* Npm (optional)

## Environment setup

Clone the Pulcy development environment tools:
```
git clone git@github.com/pulcy/pulsar.git
make
```

## Usage 

### Clearing cached data 

Pulsar keeps a cache to speed up various data fetching requests.
You can clear this cache using the following command.

```
pulsar clear cache
```

### Clone a repository 

Use the following command to clone a (git) repository into a given folder, optionally checking out a specific 
version.
The command will warn about the existance of newer versions and is very fast because it caches repositories, and makes 
use of git to fetch only missing deltas.

```
pulsar get [-b <version>] <repository-url> <folder>
```

### Fast `go get`  

Use the following command to perform a typical `go get` command, but a lot faster due to aggressive caching.
The result is fetched into `$GOPATH/src`.

```
pulsar go get <repository-url>
```

### Create local GOPATH for any go repository

Use the following command to create a local `GOPATH` structure for any local repository.
It creates a local `.gobuild/src/github.com/yourname/yourrepo` folder structure where the deepest folder
links back (via aa softlink) to the repository itself.
It then prints out the proper value for the `GOPATH` environment variable.

```
pulsar go path [-p alternative-package-name]
```

Typical use:

```
export GOPATH=$(pulsar go path)
```

### Vendor go libraries 

Use the following command to copy (vendor) one or more go libraries into a vendor folder.
If the `--flatten` argument is set, the resulting vendor directory will be flattened afterwards.

```
pulsar go vendor [-V <vendor-folder>] [--flatten] <repository>...
```

### Flattening go vendor folders  

When vendoring go libraries, the vendored libraries themselves can also hold vendored libraries.
The following command is used to copy a vendor folder to a new (temporary) directory and move all 
vendored libraries at the lowest level of that folder, or to move all vendored libraries to the 
lowest level of a target folder, without copying to another folder.

```
pulsar go flatten [-V <vendor-folder>] [<folder>] 
```

