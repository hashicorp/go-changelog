FROM golang

ENV GO111MODULE=on
WORKDIR /go-changelog

# Cache Go modules so we only download when they change
COPY go.mod .
COPY go.sum .
RUN go mod download

# Add and build the binaries
COPY . .
RUN go install ./...
