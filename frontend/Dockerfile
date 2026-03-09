# syntax=docker/dockerfile:1
# Stage 1: Build Vue app (BuildKit cache for npm)
FROM node:20-alpine AS builder
WORKDIR /app

ENV NPM_CONFIG_UPDATE_NOTIFIER=false NPM_CONFIG_FUND=false

COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
# Vite production build chỉ đọc .env.production
# Có .env.development → dùng (Docker local); không có → fallback /api (CI/prod)
RUN if [ -f .env.development ]; then cp .env.development .env.production; else echo "VITE_API_BASE_URL=/api" > .env.production; fi
RUN npm run build

# Stage 2: Nginx (minimal runtime)
FROM nginx:1.25-alpine

RUN rm -f /etc/nginx/conf.d/default.conf \
    && apk add --no-cache wget

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/conf.d/default.conf.template /etc/nginx/conf.d/default.conf.template
COPY nginx/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
