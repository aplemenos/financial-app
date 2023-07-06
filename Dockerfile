FROM golang:1.20 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/financial-app/main.go

FROM alpine:latest AS production
COPY --from=builder /app .

RUN apk --no-cache add curl

CMD ["./app"]