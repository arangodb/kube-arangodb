FROM scratch 

ADD bin/arangodb_operator /usr/bin/

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]