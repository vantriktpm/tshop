FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/product-service/cmd/api
RUN go build -o /bin/product-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/product-service /app/product-service

EXPOSE 5003

CMD ["/app/product-service"]

