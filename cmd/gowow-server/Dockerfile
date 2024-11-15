FROM golang:1.23-alpine3.20 as builder

RUN apk --no-cache add make git

COPY . /src

WORKDIR /src

RUN make build-server


FROM alpine:3.20

COPY --from=builder /src/bin/gowow-server /usr/local/bin/gowow-server
COPY --from=builder /src/assets/quotes.txt /opt/powow/quotes.txt

EXPOSE 7700

ENTRYPOINT ["/usr/local/bin/gowow-server", "--quotes-file-path", "/opt/powow/quotes.txt"]