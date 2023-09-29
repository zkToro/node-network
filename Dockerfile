# syntax=docker/dockerfile:1
FROM alpine AS base

FROM golang:1.19 as go-builder

WORKDIR /zktoro

COPY go.mod go.sum ./

RUN go mod download



COPY . /zktoro

RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /zktoro/main /zktoro/cmd/node/main.go

## GOOS not linus coz we want to run it in mac to test

From base
COPY --from=go-builder /zktoro/main /zktoro  
# name the node zktoro instead of zktoro-node
COPY 31337.json /
# CMD ["/zktoro-node"]
EXPOSE 8089 8090

# docker build --tag zktoro .
