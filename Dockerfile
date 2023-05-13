FROM golang:1.19-alpine as builder

WORKDIR /app

COPY . .

RUN  apk update && apk add build-base && go mod download && go build -o go_bot *.go

FROM alpine

WORKDIR /app

RUN apk update

ARG BOT_TOKEN
ENV TELEGRAM_APITOKEN=$BOT_TOKEN

COPY --from=builder /app/go_bot /app

RUN adduser --disabled-password --no-create-home john-doe

ENTRYPOINT ["./go_bot"]

USER john-doe
