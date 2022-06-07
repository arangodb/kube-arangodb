ARG IMAGE=scratch
FROM ${IMAGE}

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
ADD bin/${RELEASE_MODE}/linux/${TARGETARCH}/arangodb_operator_ops /usr/bin/arangodb_operator_ops

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]
