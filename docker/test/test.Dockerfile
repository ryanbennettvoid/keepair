FROM golang:1.19.4-alpine3.17

WORKDIR /root

RUN apk add build-base

RUN go install github.com/mitranim/gow@latest

ENTRYPOINT gow test -failfast ./...