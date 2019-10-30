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

ADD bin/arangodb_operator /usr/bin/

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]