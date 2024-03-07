FROM golang:1.16-alpine AS backend
WORKDIR /go/src/cedana-shell
COPY ./cmd ./cmd
COPY ./xterm ./xterm
COPY ./go.mod .
COPY ./go.sum .
ENV CGO_ENABLED=0
RUN go mod vendor
ARG VERSION_INFO=dev-build
RUN go build -a -v -o ./bin ./cmd

FROM node:16.0.0-alpine AS frontend
WORKDIR /app
COPY ./package.json .
COPY ./package-lock.json .
RUN npm install

FROM alpine:3.14.0
WORKDIR /app
RUN apk add --no-cache bash ncurses
COPY --from=backend /go/src/cedana-shell/bin /app/cedana-shell
COPY --from=frontend /app/node_modules /app/node_modules
COPY ./public /app/public
RUN ln -s /app/cedana-shell /usr/bin
RUN adduser -D -u 1000 user
RUN mkdir -p /home/user
RUN chown user:user /app -R
WORKDIR /
ENV WORKDIR=/app
USER user
ENTRYPOINT ["/app/cedana-shell"]
