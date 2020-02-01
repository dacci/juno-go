FROM golang:stretch AS builder

RUN go get github.com/dacci/juno-go

FROM debian:stretch-slim

COPY --from=builder /go/bin/juno-go /usr/local/bin
