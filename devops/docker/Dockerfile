FROM golang:1.18-alpine3.16 AS builder
RUN apk update && apk add --no-cache ca-certificates git
WORKDIR $GOPATH/src/trolly.hunterwilkins.dev/trolly/
COPY . .
RUN go get -d -v ./cmd/web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags='-w -s' -o=/go/bin/web ./cmd/web

FROM scratch
COPY --from=builder /go/bin/web /go/bin/web
COPY --from=builder /go/src/trolly.hunterwilkins.dev/trolly/ui /ui
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 80
EXPOSE 443
VOLUME [ "/certs" ]
ENTRYPOINT [ "/go/bin/web" ]