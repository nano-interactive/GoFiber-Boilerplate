FROM golang:1.18 as develop

RUN apt update && \
    apt upgrade -y && \
    apt install iputils-ping mlocate vim -y && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /bin && \
    updatedb && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest && \
    go install github.com/ramya-rao-a/go-outline@latest && \
    go install github.com/cweill/gotests/gotests@latest && \
    go install github.com/fatih/gomodifytags@latest && \
    go install github.com/josharian/impl@latest && \
    go install github.com/haya14busa/goplay/cmd/goplay@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install mvdan.cc/gofumpt@latest

WORKDIR /app

COPY . .

EXPOSE 2345
EXPOSE 80
EXPOSE 5000
EXPOSE 3000

FROM golang:1.18 as build

ARG VERSION
ARG APP_NAME

WORKDIR /app

COPY . .

RUN apt update && \
    apt install make -y && \
    make build VERSION=${VERSION} ENV=production APP_NAME=${APP_NAME}

FROM alpine:3 as production

ARG APP_NAME

WORKDIR /app

COPY --from=build /app/bin/${APP_NAME} .
COPY --from=build /app/bin/config.yml /etc/${APP_NAME}/config.yml

RUN apk update

EXPOSE 80
EXPOSE 5000
EXPOSE 3000

ENTRYPOINT [ "/app/${APP_NAME}" ]
