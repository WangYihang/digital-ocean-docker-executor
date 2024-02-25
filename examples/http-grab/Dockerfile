# syntax=docker/dockerfile:1

FROM golang:1.21.6-alpine3.19
RUN go install github.com/WangYihang/http-grab@73fa8b5aaf5c04847872f771a47b36db619ae99d
ENTRYPOINT [ "http-grab" ]