FROM golang:1.21 as build
WORKDIR /go/src/trolly
COPY go.mod go.sum tailwind.config.js cmd components internal static .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /go/bin/trolly

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/trolly /
CMD ["/trolly"]