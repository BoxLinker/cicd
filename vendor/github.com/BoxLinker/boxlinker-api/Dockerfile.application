FROM alpine:latest
RUN apk update
RUN apk add ca-certificates
RUN apk add -U tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN mkdir -p /app/config

COPY ./cmd/application/config.yml /app/config/config.yml
COPY ./cmd/application/application /app/application

CMD /app/application --config-file=/app/config/config.yml

