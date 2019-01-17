#!/bin/bash

## This script will test creating a manifest list against a specified registry
## that claims to support the Docker v2 distribution API and the v2.2 image
## specification.

## It expects `manifest-tool` to be in the $PATH as well as the docker client.
## You must be authenticated via `docker login` to the registry provided or
## whatever method that registry provides for inserting docker authentication.

## It will pull 4 images from DockerHub (alpine for 4 architectures)
## and tag them against the provided registry; push them as images to that
## registry/repo and then use the `manifest-tool` to assemble them into a
## manifest list and push using the V2 API and V2 features (like cross-repository
## push and references to blobs already existing)

_REGISTRY="${1}"

_IMAGELIST="s390x/alpine
ppc64le/alpine
aarch64/alpine
alpine"

[ -z "${_REGISTRY}" ] && {
	echo "Please provide a registry URL + namespace/repo name as the first parameter"
	exit 1
}

echo "Warning: some commands will fail if you are not authenticated to ${_REGISTRY}"

echo ">> 1: Pulling required images from DockerHub"
for i in $_IMAGELIST; do
	docker pull ${i}:latest
done

echo ">> 2: Tagging and pushing images to registry ${_REGISTRY}"
for i in $_IMAGELIST; do
	target="${i/\//_}"
	[ "${target}" == "${i}" ] && {
		# special case for no arch prefix on amd64 (x86_64 Linux) images  
		target="amd64_${i}"
	}
	echo docker tag ${i}:latest ${_REGISTRY}/${target}:latest
	docker tag ${i}:latest ${_REGISTRY}/${target}:latest
	docker push ${_REGISTRY}/${target}:latest
done

echo ">> 4: Attempt creating manifest list on registry ${_REGISTRY}"

sed s,__REGISTRY__,${_REGISTRY}, test-registry.yml >test-registry.yaml
manifest-tool --debug push from-spec test-registry.yaml

