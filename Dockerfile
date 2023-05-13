FROM golang:1.19-alpine as builder

WORKDIR /app

COPY . .

RUN  apk update && apk add build-base && go mod download && go build -o go_bot *.go

FROM alpine

WORKDIR /app

RUN apk update

RUN adduser --disabled-password --no-create-home john-doe && chown john-doe:john-doe -R /app/

COPY --from=builder /app/.env /app/go_bot /app/

ENTRYPOINT ["./go_bot"]

USER john-doe
