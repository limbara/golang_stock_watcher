FROM golang:alpine

ARG MIGRATE_VERSION=v4.15.1

RUN apk update && \
  apk add --no-cache git tzdata curl &&\
  addgroup -g 1000 author && \
  adduser -u 1000 -G author -s /bin/sh -D author

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar -xvz \
  && mv migrate /usr/local/bin

RUN curl -fLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
  && chmod +x install.sh && sh install.sh && cp ./bin/air /bin/air

WORKDIR /home/stock_watcher_server

COPY go.mod .

ENV GO111MODULE on
RUN go mod download
RUN go mod verify

USER author

CMD ["air", "-c", "./air.toml"]