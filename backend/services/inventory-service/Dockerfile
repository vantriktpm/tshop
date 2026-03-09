FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/inventory-service/cmd/server
RUN go build -o /bin/inventory-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/inventory-service /app/inventory-service

EXPOSE 8083

CMD ["/app/inventory-service"]

