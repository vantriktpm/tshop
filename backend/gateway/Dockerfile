FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

# Build gateway binary
RUN go build -o /bin/gateway .

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /bin/gateway /app/gateway

ENV PORT=5000
EXPOSE 5000

CMD ["/app/gateway"]

