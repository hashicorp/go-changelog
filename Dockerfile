# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

FROM golang:alpine as builder

ENV GO111MODULE=on
WORKDIR /go-changelog

# Cache Go modules so we only download when they change
COPY go.mod .
COPY go.sum .
RUN go mod download

# Add and build the binaries
COPY . .
RUN go install ./...

# Build a minimal image for the binaries
FROM alpine:latest
RUN apk --no-cache add ca-certificates mailcap
WORKDIR /go-changelog
COPY --from=builder /go/bin .
