#!/bin/sh
set -e
# Replace backend host/port in template (default: gateway:5000 for same-docker-network)
export BACKEND_API_HOST="${BACKEND_API_HOST:-gateway}"
export BACKEND_API_PORT="${BACKEND_API_PORT:-5000}"
envsubst '${BACKEND_API_HOST} ${BACKEND_API_PORT}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf
exec nginx -g "daemon off;"
