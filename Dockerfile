FROM golang:1.22-alpine AS server
WORKDIR /go/src/demo-server
COPY ./cmd ./cmd
COPY ./go.mod .
COPY ./go.sum .
# ENV CGO_ENABLED=0
RUN go mod vendor
RUN go build -a -v -o ./bin/shell ./cmd/shell

# cedana
FROM ghcr.io/cedana/cedana:latest AS cedana
WORKDIR /cedana

FROM node:16.0.0-alpine AS client
WORKDIR /app
COPY ./package.json .
COPY ./package-lock.json .
RUN npm install

FROM alpine:3.14.0
WORKDIR /app
RUN apk add --no-cache bash ncurses
COPY --from=server /go/src/demo-server/bin/shell /app/shell
COPY --from=client /app/node_modules /app/node_modules
COPY ./public /app/public
RUN ln -s /app/shell /usr/bin/shell
RUN adduser -D -u 1000 user
RUN mkdir -p /home/user
RUN chown user:user /app -R
WORKDIR /
ENV WORKDIR=/app
USER user
ENTRYPOINT ["/app/shell"]
