FROM golang:1.24-alpine AS builder

ENV GOWORK=off

WORKDIR /app
COPY . .

WORKDIR /app/services/promotion-service/cmd/server
RUN go build -o /bin/promotion-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/promotion-service /app/promotion-service

EXPOSE 8087

CMD ["/app/promotion-service"]

