# Nginx cho Frontend

## Chức năng

- **Reverse Proxy**: `/api/` → backend (gateway), cấu hình qua `BACKEND_API_HOST` / `BACKEND_API_PORT`.
- **Static file server**: Vue build (SPA) từ `/usr/share/nginx/html`, `try_files` fallback về `index.html`.
- **SSL termination**: Bỏ comment block `listen 443 ssl` và mount cert vào `/etc/nginx/ssl` (xem `default.conf`).
- **Security**: `X-Frame-Options`, `X-Content-Type-Options`, `X-XSS-Protection`, `Referrer-Policy`, `Permissions-Policy`; `server_tokens off`.
- **Load Balancer**: Hiện tại 1 backend (env). Để cân bằng nhiều gateway, thêm file `conf.d/upstream.conf` với nhiều `server` và dùng `proxy_pass http://backend_api/api/` trong template.

## Chạy

- **Cùng backend stack**: từ `backend/` chạy `docker-compose up -d` (đã có service `frontend`), mở http://localhost:3000.
- **Chỉ frontend**: từ `frontend/` chạy `docker-compose up -d`, set `BACKEND_API_HOST=host.docker.internal`, `BACKEND_API_PORT=5000` nếu gateway chạy trên host.

## SSL (production)

1. Đặt cert: `fullchain.pem`, `privkey.pem` vào thư mục (vd. `./ssl`).
2. Trong `default.conf` (hoặc bản copy từ template): bỏ comment `listen 443 ssl`, `ssl_certificate`, `ssl_certificate_key`, và block redirect 80→301.
3. Docker: mount volume `./ssl:/etc/nginx/ssl:ro`, expose 443.
