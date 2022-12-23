FROM golang:1.18-alpine3.16 AS builder
RUN apk update && apk add --no-cache git
WORKDIR /trolly
COPY . ./
RUN go build -ldflags='-s' -o=./bin/web ./cmd/web
ENTRYPOINT ["./bin/web"]