FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/shipping-service/cmd/server
RUN go build -o /bin/shipping-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/shipping-service /app/shipping-service

EXPOSE 8086

CMD ["/app/shipping-service"]

