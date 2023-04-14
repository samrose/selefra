FROM golang:1.20

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y postgresql postgresql-contrib

ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD pass

ENV POSTGRES_PORT 15432

ENV SELEFRA_DATABASE_DSN 'host=127.0.0.1 user=postgres password=pass port=15432 dbname=postgres sslmode=disable'

COPY . /selefra

WORKDIR /selefra

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod tidy

RUN go build -o selefra

RUN mv selefra /usr/local/bin/

EXPOSE 15432