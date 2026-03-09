FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/order-service/cmd/server
RUN go build -o /bin/order-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/order-service /app/order-service

EXPOSE 8081

CMD ["/app/order-service"]

