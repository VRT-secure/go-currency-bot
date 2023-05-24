FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN  apk update && apk add build-base

RUN go mod download

COPY . .

RUN go build -o go_bot main.go

FROM alpine

WORKDIR /app

RUN apk update && apk add --no-cache tzdata

ENV TZ=Europe/Moscow

# устанавливаем московское время
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN adduser --disabled-password --no-create-home john-doe && chown john-doe:john-doe -R /app/

COPY --from=builder /app/go_bot /app/

USER john-doe
