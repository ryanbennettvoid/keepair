FROM golang:1.19.4-alpine3.17

WORKDIR /root

RUN go install github.com/mitranim/gow@latest

ENTRYPOINT gow run cmd/master/main.go