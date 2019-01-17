FROM scratch 

ARG GOARCH=amd64

ADD bin/linux/${GOARCH}/arangodb_operator /usr/bin/

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]