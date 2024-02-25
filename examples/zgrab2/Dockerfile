# syntax=docker/dockerfile:1

FROM golang:1.21.6-alpine3.19
RUN go install github.com/zmap/zgrab2/cmd/zgrab2@latest
ENTRYPOINT [ "zgrab2" ]