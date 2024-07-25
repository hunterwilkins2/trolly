FROM golang:1.21 as build
WORKDIR /go/src/trolly
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /go/bin/trolly /go/src/trolly/cmd/web

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/trolly /
COPY --from=build /go/src/trolly/static /static
CMD ["/trolly"]