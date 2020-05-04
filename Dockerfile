FROM golang:alpine AS builder

WORKDIR /build

COPY . .
RUN go build

FROM alpine

COPY --from=builder /build/juno-go /usr/local/bin
