FROM golang:latest

WORKDIR /app

COPY ./ /app

RUN apt-get update
RUN apt install sqlite3
RUN sqlite3 test.db
RUN go mod download
RUN go build
ENTRYPOINT ./vkTest2