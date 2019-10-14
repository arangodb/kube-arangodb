# Upgrading to golang:1.13-alpine:
# It is possible once prometheus-operator will release new version (newer then 0.33.0)

# Download packages required by kube-arangodb
FROM golang:1.12.9-alpine AS downloader

# git is required by 'go mod'
RUN apk add git

WORKDIR /app

COPY go.mod .
COPY go.sum .
# It is done only once unless go.mod has been changed
RUN go mod download



# Compile Golang kube-arangodb sources with downloaded dependencies
FROM downloader AS builder
ARG VERSION
ARG COMMIT

COPY *.go /app/
COPY pkg /app/pkg
COPY dashboard/assets.go /app/dashboard/assets.go

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV GOOS=linux

RUN go build -installsuffix netgo -ldflags "-X main.projectVersion=${VERSION} -X main.projectBuild=${COMMIT}" -o /arangodb_operator



# Build the final production image with only binary file
FROM scratch

COPY --from=builder /arangodb_operator /usr/bin/arangodb_operator

ENTRYPOINT [ "/usr/bin/arangodb_operator" ]
