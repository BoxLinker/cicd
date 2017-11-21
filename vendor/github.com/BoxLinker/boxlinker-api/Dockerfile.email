FROM alpine:latest
RUN apk update
RUN apk add ca-certificates
RUN apk add -U tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY ./cmd/email/email /email

ENTRYPOINT /email -D

