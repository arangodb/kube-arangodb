ARG IMAGE=ubuntu:24.04
ARG ENVOY_IMAGE=envoyproxy/envoy:v1.31.0

# Build Steps

FROM ${ENVOY_IMAGE} AS envoy

FROM ${IMAGE} AS base

RUN apt-get update && apt-get upgrade -y && apt-get clean

FROM base

ARG VERSION
LABEL name="kube-arangodb" \
      vendor="ArangoDB" \
      version="${VERSION}" \
      release="${VERSION}" \
      summary="ArangoDB Kubernetes Oparator" \
      description="ArangoDB Kubernetes Operator" \
      maintainer="redhat@arangodb.com"

ADD ./LICENSE /licenses/LICENSE

ARG RELEASE_MODE=community
ARG TARGETARCH
ADD bin/${RELEASE_MODE}/linux/${TARGETARCH}/arangodb_operator /usr/bin/arangodb_operator
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]
