# Pulsar Project Settings

Pulsar can be configured with a `.pulcy` file that control the way
various `pulsar` command work.

This file is JSON formatted with the following structure:

```jsonc
{
    // The name of the docker image build for this project.
    // Defaults to project name.
    "image": "<your-docker-image-name>",
    "registry": "<your-docker-registry-prefix>",
    "namespace": "<your-docker-registry-namespace>",
    // If set, grunt won't be called even if there is a Gruntfile.js
    "no-grunt": <true|false>,
    // If set, a latest tag will be set of the docker image
    "tag-latest": <true|false>,
    // If set, a tag will be set to the major version of the docker image (e.g. myimage:3)
    "tag-major-version": <true|false>,
    // If set, a tag will be set to the minor version of the docker image (e.g. myimage:3.2)
    "tag-minor-version": <true|false>,
    // If set, this branch is expected (defaults to "master")
    "git-branch": "<branch-name>",
    "targets": {
        // The name of a target in a local `Makefile` used to clean
        // temporary data. Defaults to "clean".
        "clean": "<make-target-for-cleaning>",
        // The name of a target in a local `Makefile` used to build a release.
        // Empty by default.
        "release": "<make-target-for-building-a-release>"
    },
    "manifest-files": ["<additional-manifest-files>"],
    // If set, use this instead of `./vendor` as vendor directory.
    "go-vendor-dir": "<path-for-vendor-go-code>",
    // If set, creates a file with this path containing the current version
    "gradle-config-file": "<path-for-gradle-config>",
    // If set, creates a github release with given assets.
    "github-assets": [
        {
            "path": "<relative-path-of-asset-file>",
            "label": "<optional-label-of-file>"
        }
    ]
}
```