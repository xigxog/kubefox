## Build time
FROM golang:1.21 as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=off

WORKDIR /app

COPY ./ .
RUN go build -o http-srv

## Runtime
FROM alpine
COPY --from=builder /app/http-srv /
ENTRYPOINT [ "/http-srv" ]
