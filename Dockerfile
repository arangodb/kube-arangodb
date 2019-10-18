ARG IMAGE=scratch
FROM ${IMAGE}

ADD bin/arangodb_operator /usr/bin/

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]