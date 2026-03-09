FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

WORKDIR /app/services/image-service/cmd/server
RUN go build -o /bin/image-service

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/image-service /app/image-service

EXPOSE 8089

CMD ["/app/image-service"]

