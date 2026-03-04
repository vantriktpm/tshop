# TShop Backend – Domain-based Microservices

Kiến trúc: **Domain-based microservices + Event-driven + Cloud native**, DDD + Clean Architecture bên trong từng service.

## Services (Domain)

| Domain       | Service              | Port | Database    | Ghi chú                    |
|-------------|----------------------|------|-------------|----------------------------|
| User        | user-service         | 8080 | PostgreSQL  | JWT, OAuth2                |
| Product     | product-service      | 8082 | PostgreSQL  | Search: Elasticsearch      |
| Inventory   | inventory-service     | 8083 | PostgreSQL  | Consume OrderCreated (Kafka) |
| Cart        | cart-service         | 8084 | **Redis**   | Cache Aside                |
| Order       | order-service        | 8081 | PostgreSQL  | Publish OrderCreated, Saga |
| Payment     | payment-service      | 8085 | PostgreSQL  | Consume OrderCreated       |
| Shipping    | shipping-service     | 8086 | PostgreSQL  |                            |
| Promotion   | promotion-service    | 8087 | PostgreSQL  |                            |
| Notification| notification-service | 8088 | -           | Consume events, gửi mail   |

## Kiến trúc bên trong mỗi service (Clean Architecture + DDD)

```
internal/
  domain/           # Entities, Value Objects, Repository interfaces (ports)
  usecase/          # Business logic (application)
  delivery/         # REST (Gin), gRPC (internal service-to-service)
  infrastructure/   # PostgreSQL, Redis, Kafka, Elasticsearch
```

- **External**: REST (Gin), **API Gateway** (`backend/gateway`) – router chung + CORS.
- **Internal**: gRPC (có thể bật mTLS).

### API Gateway (router chung + CORS)

Chạy `backend/gateway` (port **8000** mặc định). Frontend trỏ `VITE_API_BASE_URL=http://localhost:8000/api`.

| Path prefix       | Service          | Backend URL            |
|-------------------|------------------|------------------------|
| `/api/auth`       | user-service     | http://localhost:8080  |
| `/api/orders`     | order-service    | http://localhost:8081  |
| `/api/products`   | product-service  | http://localhost:8082  |
| `/api/inventory`  | inventory-service| http://localhost:8083  |
| `/api/cart`       | cart-service     | http://localhost:8084  |
| `/api/payment`    | payment-service  | http://localhost:8085  |
| `/api/shipping`   | shipping-service | http://localhost:8086  |
| `/api/promotion`  | promotion-service| http://localhost:8087  |
| `/api/notification` | notification-service | http://localhost:8088 |
| `/api/images`     | image-service    | http://localhost:8089  |

Gateway gắn CORS (`Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`) cho mọi response và xử lý preflight OPTIONS.

## Database per service

- **PostgreSQL**: user, product, inventory, order, payment, shipping, promotion.
- **Redis**: cart, cache (product detail, inventory hot key), rate limiting.
- **Elasticsearch**: search product.

## Event-driven (Kafka)

- **order-service** tạo order → publish **OrderCreated**.
- **inventory-service** consume → trừ kho.
- **payment-service** consume → tạo payment intent.
- **notification-service** consume → gửi mail.

Saga: Choreography (Kafka) hoặc Orchestration (order-service điều phối).

## Chạy local

1. Cài Go 1.22+, Docker (cho Postgres, Redis, Kafka nếu cần).
2. **(Tùy chọn)** Chạy gateway (router chung + CORS): `cd gateway && go run .` → lắng nghe :8000. Frontend dùng `VITE_API_BASE_URL=http://localhost:8000/api`.
3. Từ thư mục `backend`:

```bash
go work sync
cd services/order-service && go build -o ../../bin/order-service ./cmd/server
cd services/product-service && go build -o ../../bin/product-service ./cmd/server
# ... tương tự các service khác
```

3. Chạy infrastructure (ví dụ):

```bash
docker-compose up -d
```

4. Chạy từng service (theo thứ tự nếu có phụ thuộc):

```bash
./bin/order-service
./bin/product-service
# ...
```

## Production

- **K8s**: namespace theo env, HPA, rolling/canary.
- **Observability**: Prometheus (metrics), Grafana (dashboard), ELK (log), Jaeger/OpenTelemetry (trace).
- **Security**: JWT, OAuth2, rate limit tại API Gateway, mTLS giữa service.
- **CI/CD**: build image → push registry → deploy (e.g. ArgoCD).
