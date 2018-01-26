FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates

FROM scratch
WORKDIR /bin
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000 9000 80 443

ADD release/boxci-server /bin/
COPY release/boxci-server.env /bin/.env

ENTRYPOINT ["./boxci-server"]