# Build from backend root context so ../pkg is available
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY pkg/ ./pkg/
COPY worker/ ./worker/
WORKDIR /app/worker
RUN go mod download
RUN go build -o /bin/worker-sync-avatar ./cmd/worker-sync-avatar

FROM alpine:3.18
COPY --from=builder /bin/worker-sync-avatar /app/worker-sync-avatar
WORKDIR /app
ENTRYPOINT ["/app/worker-sync-avatar"]
