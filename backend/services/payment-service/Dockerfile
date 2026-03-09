FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/payment-service/cmd/server
RUN go build -o /bin/payment-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/payment-service /app/payment-service

EXPOSE 8085

CMD ["/app/payment-service"]

