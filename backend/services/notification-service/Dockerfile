FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

WORKDIR /app/services/notification-service/cmd/server
RUN go build -o /bin/notification-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/notification-service /app/notification-service

EXPOSE 8088

CMD ["/app/notification-service"]

