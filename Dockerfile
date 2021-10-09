FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /go/src/otaviokr/botaviokr-twitch-bot/config
COPY . .

RUN go get -d -v ./redis ./main && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/botaviokr-config ./main

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/botaviokr-config /botaviokr-config

VOLUME [ "/config" ]
VOLUME [ "/logs" ]

ENTRYPOINT [ "/botaviokr-config" ]
