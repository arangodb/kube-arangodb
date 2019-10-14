# Download packages required by kube-arangodb
FROM golang:1.12.9-alpine AS downloader

# git is required by 'go mod'
RUN apk add git

WORKDIR /app

COPY go.mod .
COPY go.sum .
# It is done only once unless go.mod has been changed
RUN go mod download



# Compile Golang integration tests for kube-arangodb sources with downloaded dependencies
FROM downloader AS builder

COPY tests /app/tests
COPY pkg /app/pkg
COPY dashboard/assets.go /app/dashboard/assets.go

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV GOOS=linux

RUN go test -c -installsuffix netgo -o /arangodb_operator_test /app/tests



# Build the final integration test image with only binary file
FROM scratch

COPY --from=builder /arangodb_operator_test /usr/bin/arangodb_operator

ENTRYPOINT [ "/usr/bin/arangodb_operator_test" ]