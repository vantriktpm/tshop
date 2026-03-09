# Kiến trúc chuẩn cho từng service

Mỗi service theo cấu trúc thống nhất sau:

```
cmd/
   api/main.go          # Entrypoint: load env, gọi bootstrap.New(), đăng ký route, chạy server

internal/
   bootstrap/
       app.go           # Load .env rồi gọi NewPostgres(), NewRedis(), NewKafka(), NewMinio() (tùy service)
       env.go           # loadEnv() đọc file .env (cmd/.env hoặc .env) và set vào os.Getenv
       postgres.go      # Đọc DB_DSN từ env (sau khi load .env), kết nối, EnsureDatabase + Migrate
       redis.go         # Đọc REDIS_ADDR từ env, trả *redis.Client hoặc nil nếu service không dùng
       kafka.go         # Đọc KAFKA_BROKER từ env, trả producer hoặc nil
       minio.go         # Đọc MINIO_* từ env, trả client hoặc nil (chỉ image-service)

   container/
       container.go     # Chứa các dependency đã khởi tạo; tạo repo, service, handler từ chúng

   service/
       <tên>_service.go # Use case / business logic (gọi repository interface từ domain)

   repository/
       <tên>_repo.go    # Implement domain.*Repository; GORM model + Migrate(db) nếu dùng DB

   handler/
       <tên>_handler.go # HTTP handler (Gin); nhận *service.*Service, gọi service rồi trả JSON
```

## Luồng khởi động

1. `cmd/api/main.go`: gọi `bootstrap.New()`.
2. `bootstrap/app.go`: gọi `loadEnv()` (đọc `cmd/.env` hoặc `.env`), sau đó gọi `NewPostgres()`, `NewRedis()`, … (chỉ những thứ service cần).
3. Các file `postgres.go`, `redis.go`, … dùng `os.Getenv("DB_DSN")`, `os.Getenv("REDIS_ADDR")`, … (đã được set từ .env).
4. `container.New(db, redis, …)` tạo repository → service → handler.
5. `main` đăng ký route với handler và chạy server.

## Config .env

- File config: `cmd/.env` (hoặc `.env` ở thư mục gốc service).
- Docker Compose: `env_file: ./services/<service>/cmd/.env` và `environment:` override khi chạy container.
- Bootstrap luôn gọi `loadEnv()` trước khi tạo kết nối để host/port lấy từ .env khi chạy local.

## Service đã chuẩn hóa

- **cart-service**: Redis; `cmd/api`, `bootstrap`, `container`, `repository`, `service`, `handler`.
- **product-service**: Postgres; `cmd/api`, `bootstrap`, `container`, `repository` (có Migrate), `service`, `handler`.

Các service còn lại (user, order, inventory, payment, shipping, promotion, notification, image) có thể áp dụng cùng pattern; entrypoint cũ ở `cmd/server/main.go` có thể xóa sau khi chuyển xong sang `cmd/api`.
