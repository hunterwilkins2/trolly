FROM golang:1.18-alpine3.16 AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/trolly.hunterwilkins.dev/trolly/
COPY . .
RUN go get -d -v ./cmd/web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags='-w -s' -o=/go/bin/web ./cmd/web

FROM scratch
COPY --from=builder /go/bin/web /go/bin/web
COPY --from=builder /go/src/trolly.hunterwilkins.dev/trolly/ui /ui
ENTRYPOINT [ "/go/bin/web" ]