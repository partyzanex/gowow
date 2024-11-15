FROM golang:1.23-alpine3.20 as builder

RUN apk --no-cache add make git

COPY . /src

WORKDIR /src

RUN make build-client


FROM alpine:3.20

COPY --from=builder /src/bin/gowow-client /usr/local/bin/gowow-client
