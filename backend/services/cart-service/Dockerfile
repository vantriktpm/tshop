FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/cart-service/cmd/api
RUN go build -o /bin/cart-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/cart-service /app/cart-service

EXPOSE 5005

CMD ["/app/cart-service"]

