FROM alpine:latest
RUN apk update
RUN apk add ca-certificates
RUN apk add -U tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
EXPOSE 8000 9000 80 443

WORKDIR /bin

ADD release/boxci-server /bin/
COPY release/boxci-server.env /bin/.env

ENTRYPOINT ["./boxci-server"]